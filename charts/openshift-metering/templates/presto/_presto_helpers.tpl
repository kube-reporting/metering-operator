{{- define "presto-hive-catalog-properties" -}}
connector.name=hive-hadoop2
hive.allow-drop-table=true
hive.allow-rename-table=true
hive.storage-format={{ .Values.hive.spec.config.defaultFileFormat | upper }}
hive.compression-codec=SNAPPY
hive.hdfs.authentication.type=NONE
hive.metastore.authentication.type=NONE
hive.collect-column-statistics-on-write=true
hive.metastore=file
hive.metastore.catalog.dir=/var/presto-hive/
{{- end }}

{{- define "presto-jmx-catalog-properties" -}}
connector.name=jmx
{{ end }}

{{- define "presto-blackhole-catalog-properties" -}}
connector.name=blackhole
{{ end }}

{{- define "presto-memory-catalog-properties" -}}
connector.name=memory
{{ end }}

{{- define "presto-prometheus-catalog-properties" -}}
{{- if .Values.presto.spec.config.connectors.prometheus.enabled }}
{{- with .Values.presto.spec.config.connectors.prometheus -}}
connector.name=prometheus
{{- if .config.uri }}
prometheus-uri={{ .config.uri }}
{{- else }}
prometheus-uri=http://localhost:9090
{{- end }}
{{- if .config.chunkSizeDuration }}
query-chunk-size-duration={{ .config.chunkSizeDuration }}
{{- else }}
query-chunk-size-duration=1h
{{- end }}
{{- if .config.maxQueryRangeDuration }}
max-query-range-duration={{ .config.maxQueryRangeDuration }}
{{- else }}
max-query-range-duration=1d
{{- end }}
{{- if .config.cacheDuration }}
cache-duration={{ .config.cacheDuration }}
{{- else }}
cache-duration=30s
{{- end }}
{{- if or .auth.useServiceAccountToken .auth.bearerTokenFile }}
bearer-token-file={{ .auth.bearerTokenFile }}
{{ end }} {{- /* end-if */ -}}
{{ end }} {{- /* end-with */ -}}
{{ end }} {{- /* end-if-enabled */ -}}
{{ end }} {{- /* end-define */ -}}

{{- define "presto-tpcds-catalog-properties" -}}
connector.name=tpcds
{{ end }}

{{- define "presto-tpch-catalog-properties" -}}
connector.name=tpch
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
{{- if or .Values.presto.spec.config.aws.secretName .Values.presto.spec.config.aws.createSecret }}
- name: AWS_ACCESS_KEY_ID
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.presto.spec.config.aws.secretName | default "presto-aws-credentials" }}"
      key: aws-access-key-id
- name: AWS_SECRET_ACCESS_KEY
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.presto.spec.config.aws.secretName | default "presto-aws-credentials" }}"
      key: aws-secret-access-key
{{- end }}
{{- if .Values.presto.spec.config.s3Compatible.endpoint }}
- name: AWS_ACCESS_KEY_ID
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.presto.spec.config.s3Compatible.secretName | default "presto-s3-compatible-credentials" }}"
      key: aws-access-key-id
- name: AWS_SECRET_ACCESS_KEY
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.presto.spec.config.s3Compatible.secretName | default "presto-s3-compatible-credentials" }}"
      key: aws-secret-access-key
{{- end }}
{{- end }}

{{- define "presto-image" -}}
{{- if or .Values.presto.spec.image.repository .Values.presto.spec.image.tag -}}
{{- .Values.presto.spec.image.repository | default .Values.presto.spec.image.defaultRepository }}:{{ .Values.presto.spec.image.tag | default .Values.presto.spec.image.defaultTag -}}
{{- else if .Values.presto.spec.image.defaultOverride -}}
{{- .Values.presto.spec.image.defaultOverride -}}
{{- else -}}
{{-  .Values.presto.spec.image.defaultRepository }}:{{ .Values.presto.spec.image.defaultTag -}}
{{- end -}}
{{- end -}}
