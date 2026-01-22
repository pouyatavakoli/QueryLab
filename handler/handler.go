package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	"github.com/pouyatavakoli/QueryLab/db"
)

type Handler struct {
	Sandbox *db.SandboxManager
}

func NewHandler(s *db.SandboxManager) *Handler {
	slog.Info("Creating new handler", "sandbox_manager", true)
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
	start := time.Now()

	id := h.Sandbox.GenerateSessionID()
	slog.Info("Creating session", "session_id", id)

	if _, err := h.Sandbox.CreateSandbox(id); err != nil {
		slog.Error("Failed to create sandbox", "session_id", id, "error", err)
		http.Error(w, "sandbox creation failed", 500)
		return
	}

	slog.Info("Session created successfully",
		"session_id", id,
		"duration", time.Since(start),
	)

	json.NewEncoder(w).Encode(map[string]string{"session_id": id})
}

func (h *Handler) RunQuery(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode request body", "error", err)
		http.Error(w, "bad request", 400)
		return
	}

	slog.Info("Running query",
		"session_id", req.SessionID,
		"query_length", len(req.Query),
	)

	dbName, ok := h.Sandbox.GetDB(req.SessionID)
	if !ok {
		slog.Error("Invalid session ID", "session_id", req.SessionID)
		http.Error(w, "invalid session", 400)
		return
	}

	slog.Debug("Found database for session",
		"session_id", req.SessionID,
		"db_name", dbName,
	)

	cfg := h.Sandbox.Config()
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.SandboxUser,
		cfg.SandboxPassword,
		dbName,
	)

	slog.Debug("Connecting to database",
		"host", cfg.Host,
		"port", cfg.Port,
		"dbname", dbName,
		"user", cfg.SandboxUser,
	)

	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		slog.Error("Failed to open database connection",
			"session_id", req.SessionID,
			"error", err,
		)
		http.Error(w, "db connection failed", 500)
		return
	}
	defer dbConn.Close()

	rows, err := dbConn.Query(req.Query)
	if err != nil {
		slog.Error("Query execution failed",
			"session_id", req.SessionID,
			"query", req.Query,
			"error", err,
		)
		json.NewEncoder(w).Encode(QueryResponse{Error: err.Error()})
		return
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	slog.Debug("Query executed successfully",
		"session_id", req.SessionID,
		"num_columns", len(cols),
	)

	var out [][]interface{}
	rowCount := 0

	for rows.Next() {
		row := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range row {
			ptrs[i] = &row[i]
		}
		rows.Scan(ptrs...)
		out = append(out, row)
		rowCount++
	}

	slog.Info("Query completed",
		"session_id", req.SessionID,
		"num_rows", rowCount,
		"num_columns", len(cols),
		"duration", time.Since(start),
	)

	json.NewEncoder(w).Encode(QueryResponse{
		Columns: cols,
		Rows:    out,
	})
}
