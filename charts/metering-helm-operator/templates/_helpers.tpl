{{- define "operator-deployment-spec" -}}
replicas: 1
strategy:
  type: Recreate
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
    - name: {{ .Values.operator.name }}
      image: "{{ .Values.operator.image.repository }}:{{ .Values.operator.image.tag }}"
      args: ["run-operator.sh"]
      imagePullPolicy: {{ .Values.operator.image.pullPolicy }}
      env:
      - name: HELM_RELEASE_CRD_NAME
        value: "Metering"
      - name: HELM_RELEASE_CRD_API_GROUP
        value: "metering.openshift.io"
      - name: HELM_CHART_PATH
        value: "{{ required "chartPath is required" .Values.operator.chartPath }}"
      - name: ALL_NAMESPACES
        value: "{{ .Values.operator.allNamespaces }}"
{{- if .Values.operator.targetNamespaces }}
      - name: TARGET_NAMESPACES
        value: "{{ .Values.operator.targetNamespaces | join "," }}"
{{- else if .Values.operator.targetNamespacesDownwardAPIValueFrom }}
      - name: TARGET_NAMESPACES
        valueFrom:
{{ toYaml .Values.operator.targetNamespacesDownwardAPIValueFrom | indent 12 }}
{{- end }}
      - name: MY_POD_NAME
        valueFrom:
          fieldRef:
            fieldPath: metadata.name
      - name: MY_POD_NAMESPACE
        valueFrom:
          fieldRef:
            fieldPath: metadata.namespace
      - name: HELM_HOST
        value: "127.0.0.1:44134"
      - name: HELM_WAIT
        value: "false"
      - name: HELM_RECONCILE_INTERVAL_SECONDS
        value: {{ .Values.operator.reconcileIntervalSeconds | quote }}
      - name: RELEASE_HISTORY_LIMIT
        value: "3"
{{- range $index, $item := .Values.olm.imageTags }}
      - name: {{ $item.name | replace "-" "_" | upper }}
        value: "{{ $item.from.name }}"
{{- end }}
      resources:
{{ toYaml .Values.operator.resources | indent 8 }}
    - name: tiller
      image: "{{ .Values.operator.image.repository }}:{{ .Values.operator.image.tag }}"
      args: ["tiller"]
      imagePullPolicy: {{ .Values.operator.image.pullPolicy }}
      env:
      - name: TILLER_NAMESPACE
        valueFrom:
          fieldRef:
            fieldPath: metadata.namespace
      - name: TILLER_HISTORY_MAX
        value: "3"
      resources:
        requests:
          memory: "50Mi"
          cpu: "50m"
        limits:
          memory: "100Mi"
          cpu: "50m"
      livenessProbe:
        failureThreshold: 3
        httpGet:
          path: /liveness
          port: 44135
          scheme: HTTP
        initialDelaySeconds: 1
        periodSeconds: 10
        successThreshold: 1
        timeoutSeconds: 1
      readinessProbe:
        failureThreshold: 3
        httpGet:
          path: /readiness
          port: 44135
          scheme: HTTP
        initialDelaySeconds: 1
        periodSeconds: 10
        successThreshold: 1
        timeoutSeconds: 1
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
