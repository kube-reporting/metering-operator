{{- define "presto-common-env" }}
- name: HIVE_CATALOG_hive_s3_aws___access___key
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.config.awsCredentialsSecretName }}"
      key: aws-access-key-id
      optional: true
- name: HIVE_CATALOG_hive_s3_aws___secret___key
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.config.awsCredentialsSecretName }}"
      key: aws-secret-access-key
      optional: true
- name: HIVE_CATALOG_hive_metastore_uri
  valueFrom:
    configMapKeyRef:
      name: presto-common-config
      key: hive-metastore-uri
- name: PRESTO_CONF_discovery_uri
  valueFrom:
    configMapKeyRef:
      name: presto-common-config
      key: discovery-uri
- name: PRESTO_NODE_node_environment
  valueFrom:
    configMapKeyRef:
      name: presto-common-config
      key: environment
- name: PRESTO_NODE_node_id
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
{{- define "presto-env" }}
- name: PRESTO_LOG_com_facebook_presto
  valueFrom:
    configMapKeyRef:
      name: {{ . }}
      key: log-level
- name: PRESTO_CONF_task_concurrency
  valueFrom:
    configMapKeyRef:
      name: {{ . }}
      key: task-concurrency
      optional: true
- name: PRESTO_CONF_task_max___worker___threads
  valueFrom:
    configMapKeyRef:
      name: {{ . }}
      key: task-max-worker-threads
      optional: true
- name: PRESTO_CONF_task_min___drivers
  valueFrom:
    configMapKeyRef:
      name: {{ . }}
      key: task-min-drivers
      optional: true
{{- end }}

{{- define "hive-env" }}
- name: CORE_CONF_fs_defaultFS
  valueFrom:
    configMapKeyRef:
      name: hive-common-config
      key: default-fs
      optional: true
- name: CORE_CONF_fs_s3a_access_key
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.config.awsCredentialsSecretName }}"
      key: aws-access-key-id
      optional: true
- name: CORE_CONF_fs_s3a_secret_key
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.config.awsCredentialsSecretName }}"
      key: aws-secret-access-key
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
