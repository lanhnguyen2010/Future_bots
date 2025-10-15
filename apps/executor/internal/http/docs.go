package http

import (
	"embed"
	"net/http"
)

//go:embed openapi.json docs.html
var docsFS embed.FS

func serveOpenAPI(w http.ResponseWriter, r *http.Request) {
	data, err := docsFS.ReadFile("openapi.json")
	if err != nil {
		http.Error(w, "spec unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func serveDocs(w http.ResponseWriter, r *http.Request) {
	data, err := docsFS.ReadFile("docs.html")
	if err != nil {
		http.Error(w, "docs unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}
