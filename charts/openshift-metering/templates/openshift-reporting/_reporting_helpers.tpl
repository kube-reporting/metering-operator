{{- define "new-report-datasource" }}
apiVersion: metering.openshift.io/v1
kind: ReportDataSource
metadata:
  name: "{{ .name }}"
  labels:
    operator-metering: "true"
spec:
{{ toYaml .spec | indent 2 }}
{{- end }}
