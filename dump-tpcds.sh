#!/bin/bash
db=$1
file=$db.sql
>$file

B="beeline --showHeader=false --outputformat=tsv2 -n hive -u 'jdbc:hive2://localhost:2181/;serviceDiscoveryMode=zooKeeper;zooKeeperNamespace=hiveserver2?tez.queue.name=default'" 


$B -e "show create database $db"  | sed -e 's/CREATE DATABASE/CREATE DATABASE IF NOT EXISTS/' -e '$s/$/;/' >> $file

echo "use \`$db\`;" >> $file

$B -e "use $db; show tables" | while read t
do
echo "Table: $t"
echo "----------"
$B -e "use $db; show create table $t"  | sed -e 's/TABLE/TABLE IF NOT EXISTS/' -e '$s/$/;/' >> $file
done
