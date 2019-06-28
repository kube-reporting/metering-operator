{{- define "hdfs-site-xml" -}}
<configuration>
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
    <value>file:///hadoop/dfs/name</value>
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
    <value>file:///hadoop/dfs/data</value>
  </property>
  <property>
    <name>dfs.datanode.data.dir.perm</name>
    <value>{{ .Values.hadoop.spec.hdfs.config.datanodeDataDirPerms }}</value>
  </property>
  <property>
    <name>dfs.replication</name>
    <value>{{ .Values.hadoop.spec.hdfs.config.replicationFactor }}</value>
  </property>
  <property>
    <name>net.topology.script.file.name</name>
    <value>/hadoop-scripts/topology-configuration.sh</value>
  </property>
</configuration>
{{- end }}
{{- define "core-site-xml" -}}
<configuration>
  <property>
      <name>fs.defaultFS</name>
      <value>{{ .Values.hadoop.spec.config.defaultFS }}</value>
  </property>
  <property>
    <name>fs.AbstractFileSystem.wasb.Impl</name>
    <value>org.apache.hadoop.fs.azure.Wasb</value>
  </property>
  <property>
    <name>fs.azure.account.key.meteringazure.blob.core.windows.net</name>
    <value>thxMTrGElDaP5scW8XGV+/OMdZcgYvMWczXd60tMmFsgE5Vn5Lcppaj5DlPluo7qDQiIvIWk3ghwAZkkUP5ItQ==</value>
  </property>
</configuration>
{{- end }}
