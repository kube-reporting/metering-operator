#!/bin/bash
mkdir -p    /tmp
mkdir -p    /user/hive/warehouse
chmod g+w   /tmp
chmod g+w   /user/hive/warehouse

cd $HIVE_HOME/bin
./hiveserver2 --hiveconf hive.server2.enable.doAs=false
