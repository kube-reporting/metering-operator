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
{{- if or .Values.hive.spec.config.awsCredentialsSecretName .Values.hive.spec.config.createAwsCredentialsSecret }}
- name: AWS_ACCESS_KEY_ID
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.hive.spec.config.awsCredentialsSecretName | default "hive-aws-credentials" }}"
      key: aws-access-key-id
- name: AWS_SECRET_ACCESS_KEY
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.hive.spec.config.awsCredentialsSecretName | default "hive-aws-credentials" }}"
      key: aws-secret-access-key
{{- end }}
{{- end }}
