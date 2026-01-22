package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type DBConfig struct {
	Host string
	Port string

	AdminUser     string
	AdminPassword string

	SandboxUser     string
	SandboxPassword string

	BaseDB  string
	InitSQL string

	SessionTimeout time.Duration // Timeout for session cleanup
}

type SandboxManager struct {
	mu        sync.RWMutex
	sandboxes map[string]*sandboxEntry
	config    *DBConfig
}

type sandboxEntry struct {
	dbName       string
	lastActivity time.Time
}

func NewSandboxManager(cfg *DBConfig) *SandboxManager {
	if cfg.SessionTimeout == 0 {
		cfg.SessionTimeout = 1 * time.Hour // Default 1 hour
	}

	sm := &SandboxManager{
		sandboxes: make(map[string]*sandboxEntry),
		config:    cfg,
	}

	// Start cleanup goroutine
	go sm.cleanupOldSandboxes()

	return sm
}

// GetOrCreateSession gets existing session or creates a new one
// Call this on page load/refresh to renew the session
func (s *SandboxManager) GetOrCreateSession(sessionID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if session exists and is still valid
	if entry, exists := s.sandboxes[sessionID]; exists {
		// Clean up old database before creating new one
		if err := s.dropDB(entry.dbName); err != nil {
			slog.Warn("failed to drop old database", "dbName", entry.dbName, "error", err)
		}
		// Remove old entry
		delete(s.sandboxes, sessionID)
	}

	// Generate new sandbox database
	dbName := "sandbox_" + s.randomString(6)

	if err := s.createDB(dbName); err != nil {
		slog.Error("failed to create database", "dbName", dbName, "error", err)
		return "", err
	}

	if err := s.initDB(dbName); err != nil {
		slog.Error("failed to init database", "dbName", dbName, "error", err)
		// Clean up on error
		_ = s.dropDB(dbName)
		return "", err
	}

	if err := s.grantSandboxPrivileges(dbName); err != nil {
		slog.Error("failed to grant sandbox privileges", "dbName", dbName, "error", err)
		// Clean up on error
		_ = s.dropDB(dbName)
		return "", err
	}

	// Store the new sandbox
	s.sandboxes[sessionID] = &sandboxEntry{
		dbName:       dbName,
		lastActivity: time.Now(),
	}

	return dbName, nil
}

// UpdateSessionActivity updates the last activity time for a session
// Call this on user interactions to keep session alive
func (s *SandboxManager) UpdateSessionActivity(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry, exists := s.sandboxes[sessionID]; exists {
		entry.lastActivity = time.Now()
	}
}

// GetDB gets the database name for a session
func (s *SandboxManager) GetDB(sessionID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.sandboxes[sessionID]
	if !ok {
		return "", false
	}
	return entry.dbName, true
}

// CleanupSession removes a session and its database
// Call this when user explicitly logs out or exits
func (s *SandboxManager) CleanupSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.cleanupSessionLocked(sessionID)
}

// cleanupSessionLocked does the actual cleanup (assumes lock is already held)
func (s *SandboxManager) cleanupSessionLocked(sessionID string) error {
	if entry, exists := s.sandboxes[sessionID]; exists {
		if err := s.dropDB(entry.dbName); err != nil {
			slog.Warn("failed to drop database on cleanup", "dbName", entry.dbName, "error", err)
			// Continue to delete the entry even if drop fails
		}
		delete(s.sandboxes, sessionID)
		slog.Info("cleaned up session", "sessionID", sessionID, "dbName", entry.dbName)
	}
	return nil
}

// cleanupOldSandboxes periodically cleans up inactive sessions
func (s *SandboxManager) cleanupOldSandboxes() {
	ticker := time.NewTicker(1 * time.Hour) // Check every hour
	defer ticker.Stop()

	for range ticker.C {
		s.cleanupInactiveSessions()
	}
}

// cleanupInactiveSessions removes sessions that have been inactive too long
func (s *SandboxManager) cleanupInactiveSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-s.config.SessionTimeout)
	var toDelete []string

	for sessionID, entry := range s.sandboxes {
		if entry.lastActivity.Before(cutoff) {
			toDelete = append(toDelete, sessionID)
		}
	}

	for _, sessionID := range toDelete {
		_ = s.cleanupSessionLocked(sessionID)
	}

	if len(toDelete) > 0 {
		slog.Info("cleaned up inactive sessions", "count", len(toDelete))
	}
}

// GenerateSessionID generates a unique session ID
// Use this when a user first visits (e.g., set as cookie)
func (s *SandboxManager) GenerateSessionID() string {
	// Combine timestamp with random string for uniqueness
	return fmt.Sprintf("%s_%d_%s",
		s.randomString(6),
		time.Now().UnixNano(),
		s.randomString(4),
	)
}

func (s *SandboxManager) adminConn(dbName string) (*sql.DB, error) {
	conn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		s.config.Host,
		s.config.Port,
		s.config.AdminUser,
		s.config.AdminPassword,
		dbName,
	)
	return sql.Open("postgres", conn)
}

func (s *SandboxManager) createDB(name string) error {
	db, err := s.adminConn(s.config.BaseDB)
	if err != nil {
		slog.Error("failed to connect to base db", "dbName", s.config.BaseDB, "error", err)
		return err
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf(
		"CREATE DATABASE %s OWNER %s",
		name,
		s.config.AdminUser,
	))
	return err
}

func (s *SandboxManager) dropDB(name string) error {
	db, err := s.adminConn(s.config.BaseDB)
	if err != nil {
		slog.Error("failed to connect to base db", "dbName", s.config.BaseDB, "error", err)
		return err
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", name))
	return err
}

func (s *SandboxManager) initDB(name string) error {
	db, err := s.adminConn(name)
	if err != nil {
		slog.Error("failed to connect to db", "dbName", name, "error", err)
		return err
	}
	defer db.Close()

	sqlBytes, err := os.ReadFile(s.config.InitSQL)
	if err != nil {
		slog.Error("failed to read init SQL file", "file", s.config.InitSQL, "error", err)
		return err
	}

	_, err = db.Exec(string(sqlBytes))
	return err
}

func (s *SandboxManager) grantSandboxPrivileges(dbName string) error {
	db, err := s.adminConn(dbName)
	if err != nil {
		slog.Error("failed to connect to db", "dbName", dbName, "error", err)
		return err
	}
	defer db.Close()

	stmts := []string{
		fmt.Sprintf("GRANT CONNECT ON DATABASE %s TO %s", dbName, s.config.SandboxUser),
		fmt.Sprintf("GRANT USAGE ON SCHEMA public TO %s", s.config.SandboxUser),
		fmt.Sprintf("GRANT SELECT ON ALL TABLES IN SCHEMA public TO %s", s.config.SandboxUser),
		fmt.Sprintf(`
			ALTER DEFAULT PRIVILEGES IN SCHEMA public
			GRANT SELECT ON TABLES TO %s
		`, s.config.SandboxUser),
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			slog.Error("grant error", "statement", stmt, "error", err)
			return err
		}
	}
	return nil
}

func (s *SandboxManager) randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (s *SandboxManager) Config() DBConfig {
	return *s.config
}