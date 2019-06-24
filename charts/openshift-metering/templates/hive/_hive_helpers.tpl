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
{{- if or .Values.hive.spec.con***REMOVED***g.awsCredentialsSecretName .Values.hive.spec.con***REMOVED***g.createAwsCredentialsSecret }}
- name: AWS_ACCESS_KEY_ID
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.hive.spec.con***REMOVED***g.awsCredentialsSecretName | default "hive-aws-credentials" }}"
      key: aws-access-key-id
- name: AWS_SECRET_ACCESS_KEY
  valueFrom:
    secretKeyRef:
      name: "{{ .Values.hive.spec.con***REMOVED***g.awsCredentialsSecretName | default "hive-aws-credentials" }}"
      key: aws-secret-access-key
{{- end }}
{{- end }}
