#!/bin/bash

# set -x

usage() {
  echo "Usage: $0 
        Options:
         [-s]                          create new session
         [query id]                    execute a specific TPCDS query
         [query_id_from query_id_to)   execute a range of TPCDS queries
         query id range 0 to 99"
  exit
}

OPT_NEW_SESSION=0
while getopts "sh" o; do
    case "${o}" in
        s)
            OPT_NEW_SESSION=1
            ;;
        h)
            usage
            ;;
    esac
done
shift $((OPTIND-1))

FROM_QUERY_ID=${1:-0}
TO_QUERY_ID=${2:-$((FROM_QUERY_ID+1))}

MAX_SESSIONS=2

ACCESS_KEY=$(grep aws_access_key_id ~/.aws/credentials | head -n 1 | awk '{print $NF}')
SECRET_KEY=$(grep aws_secret_access_key ~/.aws/credentials | head -n 1 | awk '{print $NF}') 
code='
import org.apache.spark.sql.SparkSession
import java.nio.file.Paths

val catalog = spark.sparkContext.getConf.get("spark.sql.catalogImplementation", "in-memory")

val uri = spark.sparkContext.getConf.get("hive.metastore.uris", "")

println(s"catalog = $catalog, metastore uri = $uri")

spark.sql("show databases").show(false)
spark.sql("use tpcds_bin_partitioned_parquet_2")
spark.sql("show tables").show(false)

val myAccessKey = "'$ACCESS_KEY'"
val mySecretKey = "'$SECRET_KEY'"
var bucket = "arunmk8/"
var cfilepath = "tpcds"

val hadoopConf = sc.hadoopConfiguration
hadoopConf.set("fs.s3n.impl", "org.apache.hadoop.fs.s3native.NativeS3FileSystem")
hadoopConf.set("fs.s3n.awsAccessKeyId", myAccessKey)
hadoopConf.set("fs.s3n.awsSecretAccessKey", mySecretKey)
var dir = "s3a://" + bucket + cfilepath + "/*"
println("** Dir: $dir")

def processQuery(sqlStr: String) = {
      val start = System.nanoTime()
      spark.sql(sqlStr).show()
      val end = System.nanoTime()
      println("Time elapsed: " + (end - start) / (1000 * 1000 * 1000) + " Seconds")
}

val data = sc.wholeTextFiles(dir)
val files = data.map { case (filename, _) => filename }
files.collect().slice('${FROM_QUERY_ID}', '${TO_QUERY_ID}').foreach(filename => {
      println(s"File: $filename")
      sc.setJobDescription(Paths.get(filename).getFileName.toString)
      val sqlStr = sc.textFile(filename).collect().mkString(" ")
      println(s"SQL: $sqlStr")
      processQuery(sqlStr)
})
'

jq -nc --arg str "$code" '{"code": $str}' > /tmp/json

# portforward livy if needed
if ! nc -vz localhost 8998 2>/dev/null
then
  echo "Setting up port forward for livy pod"
  LIVY_POD=$(kubectl get pods -l "app.kubernetes.io/name=livy" -o jsonpath="{.items[0].metadata.name}")
  if [[ -n "$LIVY_POD" ]]
  then
    nohup kubectl port-forward $LIVY_POD 8998:8998 &
    sleep 5
  else
    echo "Make sure livy pod is running"
    exit 1
  fi
fi
if ! nc -vz localhost 8998 2>/dev/null
then
  echo "Port forward failed"
  exit 1
fi

# Create session if needed
CREATE_NEW_SESSION=0
NUM_SESSIONS=$(curl -s -H 'Content-Type: application/json' "http://localhost:8998/sessions"|jq '.sessions|length')
if [[ $OPT_NEW_SESSION -eq 1 ]]
then
  if [[ $NUM_SESSIONS -ge $MAX_SESSIONS ]]
  then
    echo "$NUM_SESSIONS active session(s), cannot create new session"
    exit 1
  else
    CREATE_NEW_SESSION=1
  fi
elif [[ $NUM_SESSIONS -eq 0 ]]
then
  CREATE_NEW_SESSION=1
else
  SESSION_ID=$(curl -s -H 'Content-Type: application/json' "http://localhost:8998/sessions" | jq '.sessions[-1].id')
fi

if [[ $CREATE_NEW_SESSION -eq 1 ]]
then
  echo "Creating a new session"
  SESSION_ID=$(curl -s -X POST -H 'Content-Type: application/json' -d '{"kind": "spark"}' "http://localhost:8998/sessions" | jq '.id')
  echo "Starting session: $SESSION_ID"
  state=""
  while [[ "$state" != "idle" ]]
  do
    state=$(curl -s -H 'Content-Type: application/json' "http://localhost:8998/sessions/$SESSION_ID" | jq -r '.state')
    echo -n "."
    sleep 1
  done
  echo -e "\nSession: $SESSION_ID, state: $state"
else
  echo "Using session: $SESSION_ID"
fi

# Run the statement
export STATEMENT_ID=$(curl -s -X POST -H 'Content-Type: application/json' -d@/tmp/json "http://localhost:8998/sessions/$SESSION_ID/statements" | jq '.id')
echo "Running statement id: $STATEMENT_ID"
sleep 1
state=""
while [[ "$state" != "available" ]]
do
  state=$(curl -s -H 'Content-Type: application/json' "http://localhost:8998/sessions/$SESSION_ID/statements/$STATEMENT_ID" | jq -r '.state')
  echo -n "."
  sleep 1
done

curl -s -H 'Content-Type: application/json' "http://localhost:8998/sessions/$SESSION_ID/statements/$STATEMENT_ID" | jq '.output.data|.["text/plain"] | fromjson'

