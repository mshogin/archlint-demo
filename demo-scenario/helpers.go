//go:build ignore

// Package handler provides shared HTTP helpers used across all demo steps.
// This file is NOT part of the live demo steps - it provides compilation support.
package handler

import (
	"encoding/json"
	"net/http"
	"strings"
)

// decodeJSON decodes r.Body into dst.
func decodeJSON(r *http.Request, dst interface{}) error {
	return json.NewDecoder(r.Body).Decode(dst)
}

// writeJSON encodes v as JSON and writes it with status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// pathParam extracts the last path segment from the URL.
// In a real app this would use a router (e.g. chi.URLParam).
func pathParam(r *http.Request, _ string) string {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
