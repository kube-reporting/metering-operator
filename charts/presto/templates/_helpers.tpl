{{- de***REMOVED***ne "presto-common-env" }}
- name: HIVE_CATALOG_hive_s3_aws___access___key
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.con***REMOVED***g.awsCredentialsSecretName }}"
      key: aws-access-key-id
      optional: true
- name: HIVE_CATALOG_hive_s3_aws___secret___key
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.con***REMOVED***g.awsCredentialsSecretName }}"
      key: aws-secret-access-key
      optional: true
- name: HIVE_CATALOG_hive_metastore_uri
  valueFrom:
    con***REMOVED***gMapKeyRef:
      name: presto-common-con***REMOVED***g
      key: hive-metastore-uri
- name: PRESTO_CONF_discovery_uri
  valueFrom:
    con***REMOVED***gMapKeyRef:
      name: presto-common-con***REMOVED***g
      key: discovery-uri
- name: PRESTO_NODE_node_environment
  valueFrom:
    con***REMOVED***gMapKeyRef:
      name: presto-common-con***REMOVED***g
      key: environment
- name: PRESTO_NODE_node_id
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
{{- de***REMOVED***ne "presto-env" }}
- name: PRESTO_LOG_com_facebook_presto
  valueFrom:
    con***REMOVED***gMapKeyRef:
      name: {{ . }}
      key: log-level
- name: PRESTO_CONF_task_concurrency
  valueFrom:
    con***REMOVED***gMapKeyRef:
      name: {{ . }}
      key: task-concurrency
      optional: true
- name: PRESTO_CONF_task_max___worker___threads
  valueFrom:
    con***REMOVED***gMapKeyRef:
      name: {{ . }}
      key: task-max-worker-threads
      optional: true
- name: PRESTO_CONF_task_min___drivers
  valueFrom:
    con***REMOVED***gMapKeyRef:
      name: {{ . }}
      key: task-min-drivers
      optional: true
{{- end }}

{{- de***REMOVED***ne "hive-env" }}
- name: CORE_CONF_fs_defaultFS
  valueFrom:
    con***REMOVED***gMapKeyRef:
      name: hive-common-con***REMOVED***g
      key: default-fs
      optional: true
- name: CORE_CONF_fs_s3a_access_key
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.con***REMOVED***g.awsCredentialsSecretName }}"
      key: aws-access-key-id
      optional: true
- name: CORE_CONF_fs_s3a_secret_key
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.con***REMOVED***g.awsCredentialsSecretName }}"
      key: aws-secret-access-key
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
