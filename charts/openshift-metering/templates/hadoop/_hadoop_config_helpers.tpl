{{- de***REMOVED***ne "hdfs-site-xml" -}}
<con***REMOVED***guration>
  <property>
    <name>dfs.permissions.enabled</name>
    <value>false</value>
  </property>
  <property>
    <name>dfs.webhdfs.enabled</name>
    <value>true</value>
  </property>
  <property>
    <name>dfs.namenode.name.dir</name>
    <value>***REMOVED***le:///hadoop/dfs/name</value>
  </property>
  <property>
    <name>dfs.namenode.rpc-bind-host</name>
    <value>0.0.0.0</value>
  </property>
  <property>
    <name>dfs.namenode.servicerpc-bind-host</name>
    <value>0.0.0.0</value>
  </property>
  <property>
    <name>dfs.namenode.http-bind-host</name>
    <value>0.0.0.0</value>
  </property>
  <property>
    <name>dfs.namenode.https-bind-host</name>
    <value>0.0.0.0</value>
  </property>
  <property>
    <name>dfs.client.use.datanode.hostname</name>
    <value>true</value>
  </property>
  <property>
    <name>dfs.datanode.use.datanode.hostname</name>
    <value>true</value>
  </property>
  <property>
    <name>dfs.datanode.data.dir</name>
    <value>***REMOVED***le:///hadoop/dfs/data</value>
  </property>
  <property>
    <name>dfs.datanode.data.dir.perm</name>
    <value>{{ .Values.hadoop.spec.hdfs.con***REMOVED***g.datanodeDataDirPerms }}</value>
  </property>
  <property>
    <name>dfs.replication</name>
    <value>{{ .Values.hadoop.spec.hdfs.con***REMOVED***g.replicationFactor }}</value>
  </property>
  <property>
    <name>net.topology.script.***REMOVED***le.name</name>
    <value>/hadoop-scripts/topology-con***REMOVED***guration.sh</value>
  </property>
</con***REMOVED***guration>
{{- end }}
{{- de***REMOVED***ne "core-site-xml" -}}
<con***REMOVED***guration>
  <property>
      <name>fs.defaultFS</name>
      <value>{{ .Values.hadoop.spec.con***REMOVED***g.defaultFS }}</value>
  </property>
</con***REMOVED***guration>
{{- end }}
