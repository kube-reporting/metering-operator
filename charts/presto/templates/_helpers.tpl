{{- define "presto-hive-catalog-properties" -}}
connector.name=hive-hadoop2
hive.allow-drop-table=true
hive.allow-rename-table=true
hive.storage-format={{ .Values.spec.hive.config.defaultFileFormat | upper }}
hive.compression-codec=SNAPPY
hive.hdfs.authentication.type=NONE
hive.metastore.authentication.type=NONE
hive.metastore.uri={{ .Values.spec.hive.config.metastoreURIs }}
{{- if .Values.spec.presto.config.metastoreTimeout }}
hive.metastore-timeout={{ .Values.spec.presto.config.metastoreTimeout }}
{{- end }}
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
- name: AWS_ACCESS_KEY_ID
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.spec.config.awsCredentialsSecretName }}"
      key: aws-access-key-id
      optional: true
- name: AWS_SECRET_ACCESS_KEY
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.spec.config.awsCredentialsSecretName }}"
      key: aws-secret-access-key
      optional: true
{{- end }}

{{- define "hive-env" }}
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
- name: AWS_ACCESS_KEY_ID
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.spec.config.awsCredentialsSecretName }}"
      key: aws-access-key-id
      optional: true
- name: AWS_SECRET_ACCESS_KEY
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.spec.config.awsCredentialsSecretName }}"
      key: aws-secret-access-key
      optional: true
{{- end }}
