{{- de***REMOVED***ne "hive-env" }}
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
{{- if or .Values.hive.spec.con***REMOVED***g.aws.secretName .Values.hive.spec.con***REMOVED***g.aws.createSecret }}
- name: AWS_ACCESS_KEY_ID
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.hive.spec.con***REMOVED***g.aws.secretName | default "hive-aws-credentials" }}"
      key: aws-access-key-id
- name: AWS_SECRET_ACCESS_KEY
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.hive.spec.con***REMOVED***g.aws.secretName | default "hive-aws-credentials" }}"
      key: aws-secret-access-key
{{- end }}
{{- if .Values.hive.spec.con***REMOVED***g.s3Compatible.endpoint }}
- name: AWS_ACCESS_KEY_ID
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.hive.spec.con***REMOVED***g.s3Compatible.secretName | default "hive-s3-compatible-credentials" }}"
      key: aws-access-key-id
- name: AWS_SECRET_ACCESS_KEY
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.hive.spec.con***REMOVED***g.s3Compatible.secretName | default "hive-s3-compatible-credentials" }}"
      key: aws-secret-access-key
{{- end }}
{{- end }}

{{- de***REMOVED***ne "hive-image" -}}
{{- if or .Values.hive.spec.image.repository .Values.hive.spec.image.tag -}}
{{- .Values.hive.spec.image.repository | default .Values.hive.spec.image.defaultRepository }}:{{ .Values.hive.spec.image.tag | default .Values.hive.spec.image.defaultTag -}}
{{- ***REMOVED*** if .Values.hive.spec.image.defaultOverride -}}
{{- .Values.hive.spec.image.defaultOverride -}}
{{- ***REMOVED*** -}}
{{-  .Values.hive.spec.image.defaultRepository }}:{{ .Values.hive.spec.image.defaultTag -}}
{{- end -}}
{{- end -}}
