package workshop

import (
	"agentic-layer-custom/pkg/tools"
	"encoding/json"
	"net/http"
)

// HandleToolsCatalog serves the normalized tool catalog.
func HandleToolsCatalog(w http.ResponseWriter, r *http.Request) {
	WriteCORS(w, r)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	catalog := tools.GetNormalizedToolCatalog()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(catalog)
}

func WriteCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}
