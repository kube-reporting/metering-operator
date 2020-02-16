{{- define "hadoop-image" -}}
{{- if or .Values.hadoop.spec.image.repository .Values.hadoop.spec.image.tag -}}
{{- .Values.hadoop.spec.image.repository | default .Values.hadoop.spec.image.defaultRepository }}:{{ .Values.hadoop.spec.image.tag | default .Values.hadoop.spec.image.defaultTag -}}
{{- else if .Values.hadoop.spec.image.defaultOverride -}}
{{- .Values.hadoop.spec.image.defaultOverride -}}
{{- else -}}
{{-  .Values.hadoop.spec.image.defaultRepository }}:{{ .Values.hadoop.spec.image.defaultTag -}}
{{- end -}}
{{- end -}}
