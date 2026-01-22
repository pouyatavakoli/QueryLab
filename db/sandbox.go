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
}

type SandboxManager struct {
	mu        sync.Mutex
	sandboxes map[string]string
	config    *DBConfig
}

func NewSandboxManager(cfg *DBConfig) *SandboxManager {
	return &SandboxManager{
		sandboxes: make(map[string]string),
		config:    cfg,
	}
}

func (s *SandboxManager) Config() DBConfig {
	return *s.config
}

func (s *SandboxManager) GenerateSessionID() string {
	return fmt.Sprintf("%s_%d", s.randomString(8), time.Now().Unix())
}

func (s *SandboxManager) CreateSandbox(sessionID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if old, ok := s.sandboxes[sessionID]; ok {
		_ = s.dropDB(old)
	}

	dbName := "sandbox_" + s.randomString(6)

	if err := s.createDB(dbName); err != nil {
		slog.Error("failed to create database", "dbName", dbName, "error", err)
		return "", err
	}

	if err := s.initDB(dbName); err != nil {
		slog.Error("failed to init database", "dbName", dbName, "error", err)
		return "", err
	}

	if err := s.grantSandboxPrivileges(dbName); err != nil {
		slog.Error("failed to grant sandbox privileges", "dbName", dbName, "error", err)
		return "", err
	}

	s.sandboxes[sessionID] = dbName
	return dbName, nil
}

func (s *SandboxManager) GetDB(sessionID string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	db, ok := s.sandboxes[sessionID]
	return db, ok
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
