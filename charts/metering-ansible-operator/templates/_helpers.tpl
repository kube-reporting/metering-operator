{{- de***REMOVED***ne "operator-deployment-spec" -}}
replicas: 1
strategy:
  type: RollingUpdate
selector:
  matchLabels:
    app: {{ .Values.operator.name }}
{{- if .Values.operator.labels }}
{{ toYaml .Values.operator.labels | indent 4 }}
{{- end }}
template:
  metadata:
    labels:
      app: {{ .Values.operator.name }}
{{- if .Values.operator.labels }}
{{ toYaml .Values.operator.labels | indent 6 }}
{{- end }}
{{- if .Values.operator.annotations }}
    annotations:
{{ toYaml .Values.operator.annotations | indent 6 }}
{{- end }}
  spec:
    securityContext:
      runAsNonRoot: true
    containers:
    - name: ansible
      command:
      - /opt/ansible/scripts/ansible-logs.sh
      - /tmp/ansible-operator/runner
      - stdout
      image: "{{ .Values.operator.image.repository }}:{{ .Values.operator.image.tag }}"
      imagePullPolicy: {{ .Values.operator.image.pullPolicy }}
      volumeMounts:
      - mountPath: /tmp/ansible-operator/runner
        name: runner
        readOnly: true
    - name: operator
      image: "{{ .Values.operator.image.repository }}:{{ .Values.operator.image.tag }}"
      imagePullPolicy: {{ .Values.operator.image.pullPolicy }}
      env:
      - name: OPERATOR_NAME
        value: "metering-ansible-operator"
      - name: DISABLE_OCP_FEATURES
        value: "{{ .Values.operator.disableOCPFeatures }}"
      - name: WATCH_NAMESPACE
{{- if .Values.operator.targetNamespace }}
        value: "{{ .Values.operator.targetNamespace }}"
{{- ***REMOVED*** if .Values.operator.useTargetNamespacesDownwardAPIValueFrom }}
        valueFrom:
          ***REMOVED***eldRef:
            ***REMOVED***eldPath: metadata.annotations['olm.targetNamespaces']
{{- ***REMOVED*** }}
        valueFrom:
          ***REMOVED***eldRef:
            ***REMOVED***eldPath: metadata.namespace
{{- end }}
      - name: POD_NAME
        valueFrom:
          ***REMOVED***eldRef:
            ***REMOVED***eldPath: metadata.name
{{- range $index, $item := .Values.olm.imageTags }}
      - name: {{ $item.name | replace "-" "_" | upper | printf "%s_IMAGE" }}
        value: "{{ $item.from.name }}"
{{- end }}
      volumeMounts:
      - mountPath: /tmp/ansible-operator/runner
        name: runner
      resources:
{{ toYaml .Values.operator.resources | indent 8 }}
    volumes:
      - name: runner
        emptyDir: {}
    restartPolicy: Always
    terminationGracePeriodSeconds: 30
{{- if .Values.operator.serviceAccountName}}
    serviceAccount: {{ .Values.operator.serviceAccountName }}
{{- end }}
{{- if .Values.operator.imagePullSecrets }}
    imagePullSecrets:
{{ toYaml .Values.operator.imagePullSecrets | indent 4 }}
{{- end }}
{{ end }}


{{- de***REMOVED***ne "cluster-service-version-deployment-spec" -}}
{{- $ctxCopy := merge (dict "Values" (dict "operator" (dict "useTargetNamespacesDownwardAPIValueFrom" true))) . -}}
{{ include "operator-deployment-spec" $ctxCopy }}
{{ end }}
