{{- define "presto-hive-catalog-properties" -}}
connector.name=hive-hadoop2
hive.allow-drop-table=true
hive.allow-rename-table=true
hive.storage-format={{ .Values.spec.hive.config.defaultFileFormat | upper }}
hive.compression-codec=SNAPPY
hive.hdfs.authentication.type=NONE
hive.metastore.authentication.type=NONE
hive.metastore.uri={{ .Values.spec.hive.config.metastoreURIs }}
{{- if .Values.spec.config.awsAccessKeyID }}
hive.s3.aws-access-key={{ .Values.spec.config.awsAccessKeyID }}
{{- end}}
{{- if .Values.spec.config.awsSecretAccessKey }}
hive.s3.aws-secret-key={{ .Values.spec.config.awsSecretAccessKey }}
{{- end}}
{{ end }}

{{- define "presto-jmx-catalog-properties" -}}
connector.name=jmx
{{ end }}

{{- define "presto-common-env" }}
- name: MY_NODE_ID
  valueFrom:
    fieldRef:
      fieldPath: metadata.uid
- name: MY_NODE_NAME
  valueFrom:
    fieldRef:
      fieldPath: spec.nodeName
- name: MY_POD_NAME
  valueFrom:
    fieldRef:
      fieldPath: metadata.name
- name: MY_POD_NAMESPACE
  valueFrom:
    fieldRef:
      fieldPath: metadata.namespace
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

{{- define "hive-env" }}
- name: CORE_CONF_fs_s3a_access_key
  valueFrom:
    secretKeyRef:
      name: hive-common-secrets
      key: aws-access-key-id
      optional: true
- name: CORE_CONF_fs_s3a_secret_key
  valueFrom:
    secretKeyRef:
      name: hive-common-secrets
      key: aws-secret-access-key
      optional: true
- name: CORE_CONF_fs_defaultFS
  valueFrom:
    configMapKeyRef:
      name: hive-common-config
      key: default-fs
      optional: true
- name: HIVE_SITE_CONF_hive_metastore_uris
  valueFrom:
    configMapKeyRef:
      name: hive-common-config
      key: metastore-uris
- name: HIVE_SITE_CONF_javax_jdo_option_ConnectionURL
  valueFrom:
    configMapKeyRef:
      name: hive-common-config
      key: db-connection-url
- name: HIVE_SITE_CONF_javax_jdo_option_ConnectionDriverName
  valueFrom:
    configMapKeyRef:
      name: hive-common-config
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
- name: HIVE_SITE_CONF_hive_metastore_schema_verification
  valueFrom:
    configMapKeyRef:
      name: hive-common-config
      key: enable-metastore-schema-verification
- name: HIVE_SITE_CONF_datanucleus_schema_autoCreateAll
  valueFrom:
    configMapKeyRef:
      name: hive-common-config
      key: auto-create-metastore-schema
- name: HIVE_SITE_CONF_hive_default_fileformat
  valueFrom:
    configMapKeyRef:
      name: hive-common-config
      key: default-file-format
- name: MY_NODE_NAME
  valueFrom:
    fieldRef:
      fieldPath: spec.nodeName
- name: MY_POD_NAME
  valueFrom:
    fieldRef:
      fieldPath: metadata.name
- name: MY_POD_NAMESPACE
  valueFrom:
    fieldRef:
      fieldPath: metadata.namespace
- name: JAVA_MAX_MEM_RATIO
  value: "50"
{{- end }}
