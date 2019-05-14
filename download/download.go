package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type Config struct {
	user string
	pass string
	url  string
	path string
}

type Downloader struct {
	conf   Config
	client http.Client
	sha    map[string]string
	// values for template replacement e.g. Env -> {k:v ...}
	tab map[string]map[string]string
}

func NewDownloader(userName string, password string, url string, downloadPath string) *Downloader {
	d := Downloader{
		conf: Config{
			user: userName,
			pass: password,
			url:  url,
			path: downloadPath},
		client: http.Client{}}
	shaFiles, err := filepath.Glob(path.Join(d.conf.path, "*.sha"))
	handleErr(err)
	if _, err := os.Stat(d.conf.path); os.IsNotExist(err) {
		os.MkdirAll(d.conf.path, 0755)
	}
	// load sha files
	d.sha = make(map[string]string)
	for _, shaFile := range shaFiles {
		fileName := strings.TrimSuffix(path.Base(shaFile), ".sha")
		d.sha[fileName] = readFile(shaFile)
	}
	d.tab = make(map[string]map[string]string)
	// load env
	env := make(map[string]string)
	for _, s := range os.Environ() {
		val := strings.SplitN(s, "=", 2)
		env[val[0]] = val[1]
	}
	d.tab["Env"] = env
	fmt.Println(d.conf.url, "->", d.conf.path)
	return &d
}

func (d *Downloader) CheckUpdates() {
	// GET /repos/:owner/:repo/contents/:path
	req, err := http.NewRequest(http.MethodGet, d.conf.url, nil)
	handleErr(err)
	req.SetBasicAuth(d.conf.user, d.conf.pass)
	resp, err := d.client.Do(req)
	handleErr(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	handleErr(err)
	type t map[string]interface{}
	m := make([]t, 0)
	jsonerr := json.Unmarshal(body, &m)
	handleErr(jsonerr)
	count := 0
	for _, entry := range m {
		name := fmt.Sprintf("%v", entry["name"])
		sha := fmt.Sprintf("%v", entry["sha"])
		url := fmt.Sprintf("%v", entry["download_url"])
		size, _ := strconv.ParseInt(fmt.Sprintf("%v", entry["size"]), 10, 64)
		if d.sha[name] != sha {
			destPath := path.Join(d.conf.path, name)
			fmt.Println("Downloading", url)
			tmpFile, err := ioutil.TempFile(d.conf.path, name)
			handleErr(err)
			n := d.download(url, tmpFile.Name())
			tmpFile.Close()
			// sometimes the files are cached so enforce size check
			if n == size {
				src, dest := d.mayBeParse(tmpFile.Name(), destPath)
				os.Rename(src, dest)
				fmt.Println("Downloaded to", dest)
				shaFile := path.Join(d.conf.path, name+".sha")
				fmt.Println("Writing", shaFile)
				writeFile(shaFile, sha)
				d.sha[name] = sha
				count = count + 1
			} else {
				os.Remove(tmpFile.Name())
				log.Println("WARN: Expected size:", size, "Downloaded size:", n, ", will retry.")
			}
		}
	}
	fmt.Println("Path: ", d.conf.path, ", Updated", count, "files")
}

func (d *Downloader) mayBeParse(filePath string, dest string) (string, string) {
	ext := filepath.Ext(dest)
	if ext != ".tpl" && ext != ".tmpl" {
		return filePath, dest
	}
	defer os.Remove(filePath)
	fmt.Println("Parsing template", dest)
	// the parsed template goes here
	tmpFile, err := ioutil.TempFile(filepath.Dir(filePath), "tmpl")
	defer tmpFile.Close()
	handleErr(err)
	b, err := ioutil.ReadFile(filePath)
	handleErr(err)
	t := template.New("output")
	t.Parse(string(b))
	// replace variables from tab
	t.Execute(tmpFile, d.tab)
	return tmpFile.Name(), strings.TrimSuffix(dest, ext)
}

func (d *Downloader) download(url string, filePath string) int64 {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	handleErr(err)
	req.Header.Add("Cache-Control", "no-cache")
	out, err := os.Create(filePath)
	handleErr(err)
	defer out.Close()
	resp, err := d.client.Do(req)
	handleErr(err)
	defer resp.Body.Close()
	n, err := io.Copy(out, resp.Body)
	handleErr(err)
	return n
}

func readFile(filePath string) string {
	content, err := ioutil.ReadFile(filePath)
	handleErr(err)
	return string(content)
}

func writeFile(filePath string, contents string) {
	message := []byte(contents)
	err := ioutil.WriteFile(filePath, message, 0644)
	handleErr(err)
}

func handleErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func runAt(t time.Time, fn func()) {
	fmt.Println("Scheduling at", t)
	fn()
}

func schedule(duration time.Duration, fn func()) {
	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	// once instantly
	runAt(time.Now(), fn)
	for t := range ticker.C {
		runAt(t, fn)
	}
}

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s %s %s %s %s %s\n", os.Args[0],
			"[github_user_name]", "[github_password/token]", "[github_url]", "[github_dir]", "[download_path] ...")
		fmt.Fprintf(os.Stderr, "Example: %s %s %s %s %s %s %s %s\n", os.Args[0], "arunmahadevan", "xyz",
			"https://github.com/arunmahadevan/airflow-test", "jars", "/tmp/local/jars", "hadoop-conf", "/opt/hadoop/conf")
		os.Exit(1)
	}

	d := make([]*Downloader, 0)
	userName, password, githubUrl := os.Args[1], os.Args[2], os.Args[3]
	parts := strings.Split(strings.TrimRight(githubUrl, "/"), "/")
	user, repo := parts[len(parts)-2], parts[len(parts)-1]
	for i := 4; i < len(os.Args); i += 2 {
		gitDir, downloadPath := os.Args[i], os.Args[i+1]
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", user, repo, gitDir)
		d = append(d, NewDownloader(userName, password, url, downloadPath))
	}

	fmt.Println("Started")
	// every min
	schedule(60*time.Second, func() {
		for _, downloader := range d {
			downloader.CheckUpdates()
		}
	})

	fmt.Println("Done")
}
