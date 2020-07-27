{{- define "operator-deployment-spec" -}}
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
      name: {{ .Values.operator.name }}
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
    - name: operator
      image: "{{ .Values.operator.image.repository }}:{{ .Values.operator.image.tag }}"
      imagePullPolicy: {{ .Values.operator.image.pullPolicy }}
      args:
      - "--zap-level=info"
      env:
      - name: ANSIBLE_DEBUG_LOGS
        value: "True"
      - name: ANSIBLE_VERBOSITY_METERINGCONFIG_METERING_OPENSHIFT_IO
        value: "1"
      - name: OPERATOR_NAME
        value: "{{ .Values.operator.name }}"
      - name: DISABLE_OCP_FEATURES
        value: "{{ .Values.operator.disableOCPFeatures }}"
      - name: WATCH_NAMESPACE
{{- if .Values.operator.targetNamespace }}
        value: "{{ .Values.operator.targetNamespace }}"
{{- else if .Values.operator.useTargetNamespacesDownwardAPIValueFrom }}
        valueFrom:
          fieldRef:
            fieldPath: metadata.annotations['olm.targetNamespaces']
{{- else }}
        valueFrom:
          fieldRef:
            fieldPath: metadata.namespace
{{- end }}
      - name: POD_NAME
        valueFrom:
          fieldRef:
            fieldPath: metadata.name
{{- range $index, $item := .Values.olm.imageTags }}
      - name: {{ $item.name | replace "-" "_" | upper | printf "%s_IMAGE" }}
        value: "{{ $item.from.name }}"
{{- end }}
      ports:
      - name: http-metrics
        containerPort: 8383
      - name: cr-metrics
        containerPort: 8686
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

{{- define "cluster-service-version-deployment-spec" -}}
{{- $ctxCopy := merge (dict "Values" (dict "operator" (dict "useTargetNamespacesDownwardAPIValueFrom" true))) . -}}
{{ include "operator-deployment-spec" $ctxCopy }}
{{ end }}
