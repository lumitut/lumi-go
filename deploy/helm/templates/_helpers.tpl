{{/*
Expand the name of the chart.
*/}}
{{- define "lumi-go.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "lumi-go.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "lumi-go.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "lumi-go.labels" -}}
helm.sh/chart: {{ include "lumi-go.chart" . }}
{{ include "lumi-go.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "lumi-go.selectorLabels" -}}
app.kubernetes.io/name: {{ include "lumi-go.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "lumi-go.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "lumi-go.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create database URL
*/}}
{{- define "lumi-go.databaseUrl" -}}
{{- if .Values.database.enabled }}
{{- printf "postgres://%s:$(DATABASE_PASSWORD)@%s:%d/%s?sslmode=require" .Values.database.user .Values.database.host .Values.database.port .Values.database.name }}
{{- end }}
{{- end }}

{{/*
Create Redis URL
*/}}
{{- define "lumi-go.redisUrl" -}}
{{- if .Values.redis.enabled }}
{{- if .Values.redis.existingSecret }}
{{- printf "redis://:%s@%s:%d/%d" "$(REDIS_PASSWORD)" .Values.redis.host .Values.redis.port .Values.redis.db }}
{{- else }}
{{- printf "redis://%s:%d/%d" .Values.redis.host .Values.redis.port .Values.redis.db }}
{{- end }}
{{- end }}
{{- end }}
