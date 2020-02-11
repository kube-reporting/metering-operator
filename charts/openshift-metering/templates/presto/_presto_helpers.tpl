{{- de***REMOVED***ne "presto-hive-catalog-properties" -}}
connector.name=hive-hadoop2
hive.allow-drop-table=true
hive.allow-rename-table=true
hive.storage-format={{ .Values.hive.spec.con***REMOVED***g.defaultFileFormat | upper }}
hive.compression-codec=SNAPPY
hive.hdfs.authentication.type=NONE
hive.metastore.authentication.type=NONE
hive.collect-column-statistics-on-write=true
hive.metastore=***REMOVED***le
hive.metastore.catalog.***REMOVED***le=/var/presto-hive/
{{- end }}

{{- de***REMOVED***ne "presto-jmx-catalog-properties" -}}
connector.name=jmx
{{ end }}

{{- de***REMOVED***ne "presto-blackhole-catalog-properties" -}}
connector.name=blackhole
{{ end }}

{{- de***REMOVED***ne "presto-memory-catalog-properties" -}}
connector.name=memory
{{ end }}

{{- de***REMOVED***ne "presto-prometheus-catalog-properties" -}}
{{- if .Values.presto.spec.con***REMOVED***g.connectors.prometheus.enabled }}
{{- with .Values.presto.spec.con***REMOVED***g.connectors.prometheus -}}
connector.name=prometheus
{{- if .con***REMOVED***g.uri }}
prometheus-uri={{ .con***REMOVED***g.uri }}
{{- ***REMOVED*** }}
prometheus-uri=http://localhost:9090
{{- end }}
{{- if .con***REMOVED***g.chunkSizeDuration }}
query-chunk-size-duration={{ .con***REMOVED***g.chunkSizeDuration }}
{{- ***REMOVED*** }}
query-chunk-size-duration=1h
{{- end }}
{{- if .con***REMOVED***g.maxQueryRangeDuration }}
max-query-range-duration={{ .con***REMOVED***g.maxQueryRangeDuration }}
{{- ***REMOVED*** }}
max-query-range-duration=1d
{{- end }}
{{- if .con***REMOVED***g.cacheDuration }}
cache-duration={{ .con***REMOVED***g.cacheDuration }}
{{- ***REMOVED*** }}
cache-duration=30s
{{- end }}
{{- if or .auth.useServiceAccountToken .auth.bearerTokenFile }}
bearer-token-***REMOVED***le={{ .auth.bearerTokenFile }}
{{ end }} {{- /* end-if */ -}}
{{ end }} {{- /* end-with */ -}}
{{ end }} {{- /* end-if-enabled */ -}}
{{ end }} {{- /* end-de***REMOVED***ne */ -}}

{{- de***REMOVED***ne "presto-tpcds-catalog-properties" -}}
connector.name=tpcds
{{ end }}

{{- de***REMOVED***ne "presto-tpch-catalog-properties" -}}
connector.name=tpch
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
{{- if or .Values.presto.spec.con***REMOVED***g.aws.secretName .Values.presto.spec.con***REMOVED***g.aws.createSecret }}
- name: AWS_ACCESS_KEY_ID
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.presto.spec.con***REMOVED***g.aws.secretName | default "presto-aws-credentials" }}"
      key: aws-access-key-id
- name: AWS_SECRET_ACCESS_KEY
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.presto.spec.con***REMOVED***g.aws.secretName | default "presto-aws-credentials" }}"
      key: aws-secret-access-key
{{- end }}
{{- if .Values.presto.spec.con***REMOVED***g.s3Compatible.endpoint }}
- name: AWS_ACCESS_KEY_ID
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.presto.spec.con***REMOVED***g.s3Compatible.secretName | default "presto-s3-compatible-credentials" }}"
      key: aws-access-key-id
- name: AWS_SECRET_ACCESS_KEY
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.presto.spec.con***REMOVED***g.s3Compatible.secretName | default "presto-s3-compatible-credentials" }}"
      key: aws-secret-access-key
{{- end }}
{{- end }}

{{- de***REMOVED***ne "presto-image" -}}
{{- if or .Values.presto.spec.image.repository .Values.presto.spec.image.tag -}}
{{- .Values.presto.spec.image.repository | default .Values.presto.spec.image.defaultRepository }}:{{ .Values.presto.spec.image.tag | default .Values.presto.spec.image.defaultTag -}}
{{- ***REMOVED*** if .Values.presto.spec.image.defaultOverride -}}
{{- .Values.presto.spec.image.defaultOverride -}}
{{- ***REMOVED*** -}}
{{-  .Values.presto.spec.image.defaultRepository }}:{{ .Values.presto.spec.image.defaultTag -}}
{{- end -}}
{{- end -}}
