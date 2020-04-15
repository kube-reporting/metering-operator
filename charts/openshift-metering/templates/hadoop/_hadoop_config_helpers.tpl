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
{{- if and .Values.hadoop.spec.config.aws.region .Values.useIPV6Networking }}
  <property>
    <name>fs.s3a.endpoint</name>
    <value>https://s3.dualstack.{{ .Values.hadoop.spec.config.aws.region }}.amazonaws.com</value>
  </property>
  <property>
    <name>fs.s3a.path.style.access</name>
    <value>true</value>
    <description>Enable S3 path style access.</description>
  </property>
{{- end }}
{{- if .Values.hadoop.spec.config.s3Compatible.endpoint }}
  <property>
    <name>fs.s3a.impl</name>
    <value>org.apache.hadoop.fs.s3a.S3AFileSystem</value>
    <description>The implementation of S3A Filesystem</description>
  </property>
  <property>
    <name>fs.s3a.path.style.access</name>
    <value>true</value>
    <description>Enable S3 path style access.</description>
  </property>
  <property>
    <name>fs.s3a.endpoint</name>
    <description>AWS S3 endpoint to connect to.</description>
    <value>{{ .Values.hadoop.spec.config.s3Compatible.endpoint }}</value>
  </property>
{{- end }}
  <property>
      <name>fs.gs.impl</name>
      <value>com.google.cloud.hadoop.fs.gcs.GoogleHadoopFileSystem</value>
  </property>
  <property>
      <name>fs.AbstractFileSystem.wasb.Impl</name>
      <value>org.apache.hadoop.fs.azure.Wasb</value>
  </property>
  <property>
      <name>fs.AbstractFileSystem.gs.impl</name>
      <value>com.google.cloud.hadoop.fs.gcs.GoogleHadoopFS</value>
  </property>
  <property>
      <name>fs.gs.auth.service.account.enable</name>
      <value>true</value>
  </property>
  <property>
      <name>fs.gs.reported.permissions</name>
      <value>733</value>
  </property>
{{- if and .Values.networking.useGlobalProxyNetworking .Values.presto.spec.config.tls.enabled }}
  <property>
      <name>fs.s3a.proxy.host</name>
      <value>{{ .Values.networking.proxy.config.https_proxy.hostname }}</value>
  </property>
  <property>
      <name>fs.s3a.proxy.port</name>
      <value>{{ .Values.networking.proxy.config.https_proxy.port }}</value>
  </property>
  <property>
      <name>fs.s3a.proxy.username</name>
      <value>{{ .Values.networking.proxy.config.https_proxy.username }}</value>
  </property>
  <property>
      <name>fs.s3a.proxy.password</name>
      <value>{{ .Values.networking.proxy.config.https_proxy.password }}</value>
  </property>
{{- else if .Values.networking.useGlobalProxyNetworking }}
  <property>
      <name>fs.s3a.proxy.host</name>
      <value>{{ .Values.networking.proxy.config.http_proxy.hostname }}</value>
  </property>
  <property>
      <name>fs.s3a.proxy.port</name>
      <value>{{ .Values.networking.proxy.config.http_proxy.port }}</value>
  </property>
  <property>
      <name>fs.s3a.proxy.username</name>
      <value>{{ .Values.networking.proxy.config.http_proxy.username }}</value>
  </property>
  <property>
      <name>fs.s3a.proxy.password</name>
      <value>{{ .Values.networking.proxy.config.http_proxy.password }}</value>
  </property>
{{- end }}
</configuration>
{{- end }}
