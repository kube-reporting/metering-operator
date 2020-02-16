{{- define "reporting-operator-image" -}}
{{- $operatorValues :=  index .Values "reporting-operator" -}}
{{- if or $operatorValues.spec.image.repository $operatorValues.spec.image.tag -}}
{{- $operatorValues.spec.image.repository | default $operatorValues.spec.image.defaultRepository }}:{{ $operatorValues.spec.image.tag | default $operatorValues.spec.image.defaultTag -}}
{{- else if $operatorValues.spec.image.defaultOverride -}}
{{- $operatorValues.spec.image.defaultOverride -}}
{{- else -}}
{{-  $operatorValues.spec.image.defaultRepository }}:{{ $operatorValues.spec.image.defaultTag -}}
{{- end -}}
{{- end -}}

{{- define "reporting-operator-auth-proxy-image" -}}
{{- $operatorValues :=  index .Values "reporting-operator" -}}
{{- if or $operatorValues.spec.authProxy.image.repository $operatorValues.spec.authProxy.image.tag -}}
{{- $operatorValues.spec.authProxy.image.repository | default $operatorValues.spec.authProxy.image.defaultRepository }}:{{ $operatorValues.spec.authProxy.image.tag | default $operatorValues.spec.authProxy.image.defaultTag -}}
{{- else if $operatorValues.spec.authProxy.image.defaultOverride -}}
{{- $operatorValues.spec.authProxy.image.defaultOverride -}}
{{- else -}}
{{-  $operatorValues.spec.authProxy.image.defaultRepository }}:{{ $operatorValues.spec.authProxy.image.defaultTag -}}
{{- end -}}
{{- end -}}
