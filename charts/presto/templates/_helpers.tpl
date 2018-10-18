{{- de***REMOVED***ne "presto-hive-catalog-properties" -}}
connector.name=hive-hadoop2
hive.allow-drop-table=true
hive.allow-rename-table=true
hive.storage-format={{ .Values.spec.hive.con***REMOVED***g.defaultFileFormat | upper }}
hive.compression-codec=SNAPPY
hive.hdfs.authentication.type=NONE
hive.metastore.authentication.type=NONE
hive.metastore.uri={{ .Values.spec.hive.con***REMOVED***g.metastoreURIs }}
{{- if .Values.spec.con***REMOVED***g.awsAccessKeyID }}
hive.s3.aws-access-key={{ .Values.spec.con***REMOVED***g.awsAccessKeyID }}
{{- end}}
{{- if .Values.spec.con***REMOVED***g.awsSecretAccessKey }}
hive.s3.aws-secret-key={{ .Values.spec.con***REMOVED***g.awsSecretAccessKey }}
{{- end}}
{{ end }}

{{- de***REMOVED***ne "presto-jmx-catalog-properties" -}}
connector.name=jmx
{{ end }}

{{- de***REMOVED***ne "presto-common-env" }}
- name: MY_NODE_ID
  valueFrom:
    ***REMOVED***eldRef:
      ***REMOVED***eldPath: metadata.uid
- name: MY_NODE_NAME
  valueFrom:
    ***REMOVED***eldRef:
      ***REMOVED***eldPath: spec.nodeName
- name: MY_POD_NAME
  valueFrom:
    ***REMOVED***eldRef:
      ***REMOVED***eldPath: metadata.name
- name: MY_POD_NAMESPACE
  valueFrom:
    ***REMOVED***eldRef:
      ***REMOVED***eldPath: metadata.namespace
- name: MY_MEM_REQUEST
  valueFrom:
    resourceFieldRef:
      containerName: presto
      resource: requests.memory
- name: MY_MEM_LIMIT
  valueFrom:
    resourceFieldRef:
      containerName: presto
      resource: limits.memory
- name: JAVA_MAX_MEM_RATIO
  value: "50"
{{- end }}

{{- de***REMOVED***ne "hive-env" }}
- name: CORE_CONF_fs_defaultFS
  valueFrom:
    con***REMOVED***gMapKeyRef:
      name: hive-common-con***REMOVED***g
      key: default-fs
      optional: true
- name: HIVE_SITE_CONF_hive_metastore_uris
  valueFrom:
    con***REMOVED***gMapKeyRef:
      name: hive-common-con***REMOVED***g
      key: metastore-uris
- name: HIVE_SITE_CONF_javax_jdo_option_ConnectionURL
  valueFrom:
    con***REMOVED***gMapKeyRef:
      name: hive-common-con***REMOVED***g
      key: db-connection-url
- name: HIVE_SITE_CONF_javax_jdo_option_ConnectionDriverName
  valueFrom:
    con***REMOVED***gMapKeyRef:
      name: hive-common-con***REMOVED***g
      key: db-connection-driver
- name: HIVE_SITE_CONF_javax_jdo_option_ConnectionUserName
  valueFrom:
    secretKeyRef:
      name: hive-common-secrets
      key: db-connection-username
      optional: true
- name: HIVE_SITE_CONF_javax_jdo_option_ConnectionPassword
  valueFrom:
    secretKeyRef:
      name: hive-common-secrets
      key: db-connection-password
      optional: true
- name: HIVE_SITE_CONF_hive_metastore_schema_veri***REMOVED***cation
  valueFrom:
    con***REMOVED***gMapKeyRef:
      name: hive-common-con***REMOVED***g
      key: enable-metastore-schema-veri***REMOVED***cation
- name: HIVE_SITE_CONF_datanucleus_schema_autoCreateAll
  valueFrom:
    con***REMOVED***gMapKeyRef:
      name: hive-common-con***REMOVED***g
      key: auto-create-metastore-schema
- name: HIVE_SITE_CONF_hive_default_***REMOVED***leformat
  valueFrom:
    con***REMOVED***gMapKeyRef:
      name: hive-common-con***REMOVED***g
      key: default-***REMOVED***le-format
- name: MY_NODE_NAME
  valueFrom:
    ***REMOVED***eldRef:
      ***REMOVED***eldPath: spec.nodeName
- name: MY_POD_NAME
  valueFrom:
    ***REMOVED***eldRef:
      ***REMOVED***eldPath: metadata.name
- name: MY_POD_NAMESPACE
  valueFrom:
    ***REMOVED***eldRef:
      ***REMOVED***eldPath: metadata.namespace
- name: JAVA_MAX_MEM_RATIO
  value: "50"
{{- end }}
