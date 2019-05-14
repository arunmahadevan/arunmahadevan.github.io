package com.example.cluster

import org.apache.spark.sql.SparkSession
import java.nio.file.Paths

// TPCDS test app
object TcpdsTest {

  def main(args: Array[String]) {
     val spark = SparkSession
      .builder
      .appName("Tpcds test")
      .enableHiveSupport()
      .getOrCreate()

    val catalog = spark.sparkContext.getConf.get("spark.sql.catalogImplementation", "in-memory")

    val uri = spark.sparkContext.getConf.get("hive.metastore.uris", "")

    println(s"catalog = $catalog, metastore uri = $uri")

    val db = if (args.length >= 3) args(2) else "tpcds_bin_partitioned_parquet_2"

    spark.sql("show databases").show(false)
    spark.sql(s"use $db")
    spark.sql("show tables").show(false)

    val myAccessKey = "***"
    val mySecretKey = "***"
    var bucket = "arunmk8/"
    var cfilepath = "tpcds"

    val hadoopConf = spark.sparkContext.hadoopConfiguration
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

    val data = spark.sparkContext.wholeTextFiles(dir)
    val files = data.map { case (filename, _) => filename }
    val from = if (args.length >= 1) Math.min(Math.max(args(0).toInt, 0), 98) else 0
    val to = if (args.length >= 2) Math.min(Math.max(args(1).toInt, 1), 99) else 1
    files.collect().slice(from, to).foreach(filename => {
      spark.sparkContext.setJobDescription(Paths.get(filename).getFileName.toString)
      val sqlStr = spark.sparkContext.textFile(filename).collect().mkString(" ")
      println(s"SQL: $sqlStr")
      processQuery(sqlStr)
      println(s"From: $from, To: $to")
      println(s"File: $filename")
    })

    spark.stop()
  }
}
