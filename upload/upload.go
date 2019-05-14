package main

import (
     "os"
    "fmt"
    "io/ioutil"
    "net/http"
)

func uploadFile(w http.ResponseWriter, r *http.Request) {
    fmt.Println("File Upload Endpoint Hit")
    r.ParseMultipartForm(10 << 20)
    file, handler, err := r.FormFile("dag")
    if err != nil {
        fmt.Println("Error Retrieving the File")
        fmt.Println(err)
        return
    }
    defer file.Close()
    fmt.Printf("Uploaded File: %+v\n", handler.Filename)
    fmt.Printf("File Size: %+v\n", handler.Size)
    fmt.Printf("MIME Header: %+v\n", handler.Header)
    fileBytes, err := ioutil.ReadAll(file)
    if err != nil {
        fmt.Println(err)
    }
    var dir = "/tmp"
    if (len(os.Args) >= 2) {
      dir = os.Args[1]
    }
    var dest = dir + "/" + handler.Filename
    ioutil.WriteFile(dest, fileBytes, 0755)
    msg := fmt.Sprintf("Successfully Uploaded File to: %+v\n", dest)
    fmt.Printf(msg)
    fmt.Fprintf(w, msg)
}

func setupRoutes() {
    http.HandleFunc("/upload", uploadFile)
    http.ListenAndServe(":8081", nil)
}

func main() {
   fmt.Println("Started")
    setupRoutes()
}
