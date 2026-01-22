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
	Query string `json:"query"` // Removed SessionID from request - now from cookie
}

type QueryResponse struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Error   string          `json:"error,omitempty"`
}

type SessionResponse struct {
	SessionID string `json:"session_id"`
	Success   bool   `json:"success"`
}

// getSessionIDFromCookie extracts session ID from cookies
func (h *Handler) getSessionIDFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("querylab_session")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// setSessionCookie sets the session cookie
func (h *Handler) setSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "querylab_session",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   int(h.Sandbox.Config().SessionTimeout.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		// Secure:   true, // Uncomment in production with HTTPS
	})
}

// clearSessionCookie clears the session cookie
func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "querylab_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Check if session already exists in cookie
	if existingSessionID, err := h.getSessionIDFromCookie(r); err == nil {
		slog.Info("Found existing session in cookie, refreshing",
			"session_id", existingSessionID,
		)

		// Clean up old sandbox and create new one
		dbName, err := h.Sandbox.GetOrCreateSession(existingSessionID)
		if err != nil {
			slog.Error("Failed to refresh sandbox",
				"session_id", existingSessionID,
				"error", err,
			)
			// Generate new session if refresh fails
			h.createNewSession(w, start)
			return
		}

		slog.Info("Session refreshed successfully",
			"session_id", existingSessionID,
			"db_name", dbName,
			"duration", time.Since(start),
		)

		// Refresh cookie
		h.setSessionCookie(w, existingSessionID)
		json.NewEncoder(w).Encode(SessionResponse{
			SessionID: existingSessionID,
			Success:   true,
		})
		return
	}

	// No existing session, create new one
	h.createNewSession(w, start)
}

// createNewSession creates a brand new session
func (h *Handler) createNewSession(w http.ResponseWriter, start time.Time) {
	id := h.Sandbox.GenerateSessionID()
	slog.Info("Creating new session", "session_id", id)

	dbName, err := h.Sandbox.GetOrCreateSession(id)
	if err != nil {
		slog.Error("Failed to create sandbox", "session_id", id, "error", err)
		http.Error(w, "sandbox creation failed", 500)
		return
	}

	// Set session cookie
	h.setSessionCookie(w, id)

	slog.Info("Session created successfully",
		"session_id", id,
		"db_name", dbName,
		"duration", time.Since(start),
	)

	json.NewEncoder(w).Encode(SessionResponse{
		SessionID: id,
		Success:   true,
	})
}

func (h *Handler) RunQuery(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Get session ID from cookie
	sessionID, err := h.getSessionIDFromCookie(r)
	if err != nil {
		slog.Error("No session cookie found", "error", err)
		http.Error(w, "session required", http.StatusUnauthorized)
		return
	}

	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode request body", "error", err)
		http.Error(w, "bad request", 400)
		return
	}

	slog.Info("Running query",
		"session_id", sessionID,
		"query_length", len(req.Query),
	)

	// Get or create sandbox for this session
	dbName, err := h.Sandbox.GetOrCreateSession(sessionID)
	if err != nil {
		slog.Error("Failed to get/create sandbox",
			"session_id", sessionID,
			"error", err,
		)
		http.Error(w, "sandbox error", 500)
		return
	}

	// Update session activity to keep it alive
	h.Sandbox.UpdateSessionActivity(sessionID)

	slog.Debug("Found database for session",
		"session_id", sessionID,
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
			"session_id", sessionID,
			"error", err,
		)
		http.Error(w, "db connection failed", 500)
		return
	}
	defer dbConn.Close()

	rows, err := dbConn.Query(req.Query)
	if err != nil {
		slog.Error("Query execution failed",
			"session_id", sessionID,
			"query", req.Query,
			"error", err,
		)
		json.NewEncoder(w).Encode(QueryResponse{Error: err.Error()})
		return
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	slog.Debug("Query executed successfully",
		"session_id", sessionID,
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

	if err := rows.Scan(ptrs...); err != nil {
		slog.Error("Row scan failed", "error", err)
		continue
	}

	// Normalize types for JSON
	for i, v := range row {
		switch val := v.(type) {
		case []byte:
			row[i] = string(val)
		default:
			row[i] = val
		}
	}

	out = append(out, row)
	rowCount++
}


	slog.Info("Query completed",
		"session_id", sessionID,
		"num_rows", rowCount,
		"num_columns", len(cols),
		"duration", time.Since(start),
	)

	json.NewEncoder(w).Encode(QueryResponse{
		Columns: cols,
		Rows:    out,
	})
}

// Logout handles session termination
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionID, err := h.getSessionIDFromCookie(r)
	if err != nil {
		// No session to logout from
		w.WriteHeader(http.StatusOK)
		return
	}

	// Clean up sandbox database
	if err := h.Sandbox.CleanupSession(sessionID); err != nil {
		slog.Warn("Failed to cleanup session on logout",
			"session_id", sessionID,
			"error", err,
		)
	}

	// Clear cookie
	h.clearSessionCookie(w)

	slog.Info("User logged out", "session_id", sessionID)
	w.WriteHeader(http.StatusOK)
}

// HealthCheck provides a simple health endpoint
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"time":    time.Now().Format(time.RFC3339),
		"service": "QueryLab",
	})
}
