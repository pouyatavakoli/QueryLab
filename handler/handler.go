package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
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

func (h *Handler) CreateSession(w http.ResponseWriter, _ *http.Request) {
	id := h.Sandbox.GenerateSessionID()
	if _, err := h.Sandbox.CreateSandbox(id); err != nil {
		http.Error(w, "sandbox creation failed", 500)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"session_id": id})
}

func (h *Handler) RunQuery(w http.ResponseWriter, r *http.Request) {
	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", 400)
		return
	}

	dbName, ok := h.Sandbox.GetDB(req.SessionID)
	if !ok {
		http.Error(w, "invalid session", 400)
		return
	}

	cfg := h.Sandbox.Config()
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.SandboxUser,
		cfg.SandboxPassword,
		dbName,
	)

	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		http.Error(w, "db connection failed", 500)
		return
	}
	defer dbConn.Close()

	rows, err := dbConn.Query(req.Query)
	if err != nil {
		json.NewEncoder(w).Encode(QueryResponse{Error: err.Error()})
		return
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	var out [][]interface{}

	for rows.Next() {
		row := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range row {
			ptrs[i] = &row[i]
		}
		rows.Scan(ptrs...)
		out = append(out, row)
	}

	json.NewEncoder(w).Encode(QueryResponse{
		Columns: cols,
		Rows:    out,
	})
}
