package bots

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// ManifestWriter persists Kubernetes manifests for bots.
type ManifestWriter interface {
	Write(ctx context.Context, bot Bot) (string, error)
}

// FileManifestWriter renders bot manifests into files on disk.
type FileManifestWriter struct {
	baseDir string
}

// NewFileManifestWriter creates a writer that stores manifests inside baseDir.
func NewFileManifestWriter(baseDir string) *FileManifestWriter {
	return &FileManifestWriter{baseDir: baseDir}
}

// Write renders a manifest for the given bot, creating directories as needed.
func (w *FileManifestWriter) Write(_ context.Context, bot Bot) (string, error) {
	if err := os.MkdirAll(w.baseDir, 0o755); err != nil {
		return "", fmt.Errorf("create manifest dir: %w", err)
	}
	filename := fmt.Sprintf("%s.yaml", sanitizeName(bot.ID))
	path := filepath.Join(w.baseDir, filename)

	content, err := renderManifest(bot)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return "", fmt.Errorf("write manifest: %w", err)
	}
	return path, nil
}

func renderManifest(bot Bot) ([]byte, error) {
	configJSON := "{}"
	if len(bot.Config) > 0 {
		var buf bytes.Buffer
		if err := json.Indent(&buf, bot.Config, "", "  "); err != nil {
			return nil, fmt.Errorf("pretty config: %w", err)
		}
		configJSON = buf.String()
	}

	data := struct {
		Bot              Bot
		SafeName         string
		ConfigJSON       string
		Replicas         int
		ContainerPort    int
		HeartbeatPort    int
		PrometheusPort   int
		TerminationGrace int
	}{
		Bot:              bot,
		SafeName:         sanitizeName(bot.ID),
		ConfigJSON:       configJSON,
		Replicas:         boolToInt(bot.Enabled),
		ContainerPort:    8080,
		HeartbeatPort:    8081,
		PrometheusPort:   8081,
		TerminationGrace: 5,
	}

	tpl, err := template.New("manifest").Funcs(template.FuncMap{
		"indent": indent,
	}).Parse(manifestTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render template: %w", err)
	}
	return buf.Bytes(), nil
}

func sanitizeName(id string) string {
	kubeName := strings.ToLower(id)
	kubeName = strings.ReplaceAll(kubeName, "_", "-")
	reg := regexp.MustCompile(`[^a-z0-9\-]+`)
	kubeName = reg.ReplaceAllString(kubeName, "-")
	kubeName = strings.Trim(kubeName, "-")
	if kubeName == "" {
		kubeName = "bot"
	}
	return kubeName
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func indent(spaces int, v string) string {
	padding := strings.Repeat(" ", spaces)
	lines := strings.Split(v, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lines[i] = padding + line
	}
	return strings.Join(lines, "\n")
}

const manifestTemplate = `---
apiVersion: v1
kind: Secret
metadata:
  name: bot-{{ .SafeName }}-config
  labels:
    app: bot
    bot_id: "{{ .Bot.ID }}"
type: Opaque
stringData:
  config.json: |-
{{ indent 4 .ConfigJSON }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bot-{{ .SafeName }}
  labels:
    app: bot
    bot_id: "{{ .Bot.ID }}"
    account: "{{ .Bot.AccountID }}"
spec:
  replicas: {{ .Replicas }}
  selector:
    matchLabels:
      app: bot
      bot_id: "{{ .Bot.ID }}"
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: bot
        bot_id: "{{ .Bot.ID }}"
        account: "{{ .Bot.AccountID }}"
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "{{ .PrometheusPort }}"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: bot-runner
      containers:
      - name: bot
        image: {{ .Bot.Image }}
        imagePullPolicy: IfNotPresent
        env:
        - name: BOT_ID
          value: "{{ .Bot.ID }}"
        - name: ACCOUNT_ID
          value: "{{ .Bot.AccountID }}"
        - name: BOT_NAME
          value: "{{ .Bot.Name }}"
        - name: BOT_CONFIG
          valueFrom:
            secretKeyRef:
              name: bot-{{ .SafeName }}-config
              key: config.json
        ports:
        - containerPort: {{ .HeartbeatPort }}
          name: http
        resources:
          requests:
            cpu: "200m"
            memory: "256Mi"
          limits:
            cpu: "1"
            memory: "1Gi"
        readinessProbe:
          httpGet:
            path: /readyz
            port: {{ .HeartbeatPort }}
          periodSeconds: 5
        livenessProbe:
          httpGet:
            path: /healthz
            port: {{ .HeartbeatPort }}
          periodSeconds: 5
      terminationGracePeriodSeconds: {{ .TerminationGrace }}
`
