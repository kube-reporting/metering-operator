{{- define "ghostunnel-image" -}}
{{- if or .Values.__ghostunnel.image.repository .Values.__ghostunnel.image.tag -}}
{{- .Values.__ghostunnel.image.repository | default .Values.__ghostunnel.image.defaultRepository }}:{{ .Values.__ghostunnel.image.tag | default .Values.__ghostunnel.image.defaultTag -}}
{{- else if .Values.__ghostunnel.image.defaultOverride -}}
{{- .Values.__ghostunnel.image.defaultOverride -}}
{{- else -}}
{{- .Values.__ghostunnel.image.defaultRepository }}:{{ .Values.__ghostunnel.image.defaultTag -}}
{{- end -}}
{{- end -}}
