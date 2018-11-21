{{/* vim: set ***REMOVED***letype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- de***REMOVED***ne "superset.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuf***REMOVED***x "-" -}}
{{- end -}}

{{/*
Create a default fully quali***REMOVED***ed app name.
We truncate at 63 chars because some Kubernetes name ***REMOVED***elds are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- de***REMOVED***ne "superset.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuf***REMOVED***x "-" -}}
{{- ***REMOVED*** -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuf***REMOVED***x "-" -}}
{{- ***REMOVED*** -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuf***REMOVED***x "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- de***REMOVED***ne "superset.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuf***REMOVED***x "-" -}}
{{- end -}}