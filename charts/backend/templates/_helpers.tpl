{{- /* helpers for backend chart */ -}}
{{- define "labels" -}}
app: "{{ .Release.Name }}-{{ .Chart.Name }}"
app.kubernetes.io/name: {{ .Chart.Name }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/app: {{ .Release.Name }}
app.kubernetes.io/name: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/part-of: {{ .Chart.Name }}
app.kubernetes.io/component: {{ .Chart.Name }}
app.kubernetes.io/image: "{{ .Values.image.tag }}"
{{- end }}

{{- define "annotations" -}}
meta.helm.sh/release-name: {{ .Release.Name }}
alloy.io/scrape: "true"
loki.io/logs: "true"
prometheus.io/scrape: "true"
prometheus.io/port: {{ .Values.service.listenPort | default "8080" | quote }}
prometheus.io/path: "/metrics"
{{- end }}