package db

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type SandboxManager struct {
	mu        sync.Mutex
	sandboxes map[string]string
	config    *DBConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	BaseDB   string
	InitSQL  string
}

func (s *SandboxManager) Config() DBConfig {
	return *s.config
}

func NewSandboxManager(cfg *DBConfig) *SandboxManager {
	return &SandboxManager{
		sandboxes: make(map[string]string),
		config:    cfg,
	}
}

func (s *SandboxManager) randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// Generate session ID: random string + timestamp
func (s *SandboxManager) GenerateSessionID() string {
	return fmt.Sprintf("%s_%d", s.randomString(8), time.Now().Unix())
}

func (s *SandboxManager) CreateSandbox(sessionID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Drop old sandbox
	if oldDB, ok := s.sandboxes[sessionID]; ok {
		if err := s.dropDB(oldDB); err != nil {
			log.Println("failed to drop old sandbox:", err)
		}
	}

	dbName := fmt.Sprintf("sandbox_%s", s.randomString(6))
	if err := s.createDB(dbName); err != nil {
		return "", err
	}

	if err := s.initDB(dbName); err != nil {
		return "", err
	}

	s.sandboxes[sessionID] = dbName
	return dbName, nil
}

func (s *SandboxManager) GetDB(sessionID string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	dbName, ok := s.sandboxes[sessionID]
	return dbName, ok
}

func (s *SandboxManager) createDB(name string) error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		s.config.Host, s.config.Port, s.config.User, s.config.Password, s.config.BaseDB)
	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	_, err = dbConn.Exec(fmt.Sprintf("CREATE DATABASE %s;", name))
	return err
}

func (s *SandboxManager) dropDB(name string) error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		s.config.Host, s.config.Port, s.config.User, s.config.Password, s.config.BaseDB)
	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	_, err = dbConn.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s;", name))
	return err
}

func (s *SandboxManager) initDB(name string) error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		s.config.Host, s.config.Port, s.config.User, s.config.Password, name)
	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	sqlBytes, err := os.ReadFile(s.config.InitSQL)
	if err != nil {
		return fmt.Errorf("failed to read init.sql: %w", err)
	}

	_, err = dbConn.Exec(string(sqlBytes))
	if err != nil {
		return fmt.Errorf("failed to initialize sandbox: %w", err)
	}

	return nil
}
