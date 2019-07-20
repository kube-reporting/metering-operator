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
{{- if or .Values.hive.spec.config.aws.secretName .Values.hive.spec.config.aws.createSecret }}
- name: AWS_ACCESS_KEY_ID
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.hive.spec.config.aws.secretName | default "hive-aws-credentials" }}"
      key: aws-access-key-id
- name: AWS_SECRET_ACCESS_KEY
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.hive.spec.config.aws.secretName | default "hive-aws-credentials" }}"
      key: aws-secret-access-key
{{- end }}
{{- if .Values.hive.spec.config.s3Compatible.endpoint }}
- name: AWS_ACCESS_KEY_ID
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.hive.spec.config.s3Compatible.secretName | default "hive-s3-compatible-credentials" }}"
      key: aws-access-key-id
- name: AWS_SECRET_ACCESS_KEY
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.hive.spec.config.s3Compatible.secretName | default "hive-s3-compatible-credentials" }}"
      key: aws-secret-access-key
{{- end }}
{{- end }}
