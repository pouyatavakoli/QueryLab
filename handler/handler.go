package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/pouyatavakoli/QueryLab/db"
)

type Handler struct {
	Sandbox *db.SandboxManager
}

func NewHandler(s *db.SandboxManager) *Handler {
	return &Handler{Sandbox: s}
}

type QueryRequest struct {
	SessionID string `json:"session_id"`
	Query     string `json:"query"`
}

type QueryResponse struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Error   string          `json:"error,omitempty"`
}

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	sessionID := h.Sandbox.GenerateSessionID()
	_, err := h.Sandbox.CreateSandbox(sessionID)
	if err != nil {
		http.Error(w, "Failed to create sandbox", http.StatusInternalServerError)
		return
	}
	resp := map[string]string{"session_id": sessionID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) RunQuery(w http.ResponseWriter, r *http.Request) {
	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	dbName, ok := h.Sandbox.GetDB(req.SessionID)
	if !ok {
		http.Error(w, "invalid session", http.StatusBadRequest)
		return
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		h.Sandbox.Config().Host,
		h.Sandbox.Config().Port,
		h.Sandbox.Config().User,
		h.Sandbox.Config().Password,
		dbName,
	)
	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		http.Error(w, "db connection failed", http.StatusInternalServerError)
		return
	}
	defer dbConn.Close()

	rows, err := dbConn.Query(req.Query)
	if err != nil {
		resp := QueryResponse{Error: err.Error()}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		http.Error(w, "failed to get columns", http.StatusInternalServerError)
		return
	}

	var results [][]interface{}
	for rows.Next() {
		colData := make([]interface{}, len(cols))
		colPtrs := make([]interface{}, len(cols))
		for i := range colData {
			colPtrs[i] = &colData[i]
		}
		if err := rows.Scan(colPtrs...); err != nil {
			log.Println("scan error:", err)
			continue
		}
		results = append(results, colData)
	}

	resp := QueryResponse{
		Columns: cols,
		Rows:    results,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
