#!/bin/bash -ex
mkdir -p    /tmp
mkdir -p    /user/hive/warehouse
chmod g+w   /tmp
chmod g+w   /user/hive/warehouse

export HADOOP_CLASSPATH="/opt/hive/hcatalog/share/hcatalog/*:/opt/hadoop-2.8.0/share/hadoop/tools/lib/*"
export HIVE_AUX_JARS_PATH=/usr/hdp/current/hive-server2/auxlib

cd $HIVE_HOME/bin
./hiveserver2 --hiveconf hive.server2.enable.doAs=false
