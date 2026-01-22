package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"log/slog"
	"net/http"

	"github.com/pouyatavakoli/QueryLab/config"
	"github.com/pouyatavakoli/QueryLab/db"
	"github.com/pouyatavakoli/QueryLab/handler"

	_ "github.com/lib/pq"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("starting QueryLab server",
		"timestamp", time.Now().Format(time.RFC3339),
		"pid", os.Getpid(),
	)

	cfg := config.LoadConfig()

	slog.Info("initializing sandbox database manager")
	sandbox := db.NewSandboxManager(&db.DBConfig{
		Host: cfg.DBHost,
		Port: cfg.DBPort,

		AdminUser:     cfg.DBAdminUser,
		AdminPassword: cfg.DBAdminPassword,

		SandboxUser:     cfg.DBSandboxUser,
		SandboxPassword: cfg.DBSandboxPassword,

		BaseDB:         cfg.DBName,
		InitSQL:        cfg.InitSQL,
		SessionTimeout: 1 * time.Hour,
	})
	slog.Info("sandbox database manager initialized")

	h := handler.NewHandler(sandbox)

	// Routes
	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/api/session", h.CreateSession)
	http.HandleFunc("/api/query", h.RunQuery)
	http.HandleFunc("/api/logout", h.Logout)
	http.HandleFunc("/api/health", h.HealthCheck)

	// lite frontend version
	http.Handle("/lite/", http.StripPrefix("/lite/", http.FileServer(http.Dir("frontend/lite"))))

	addr := ":" + cfg.ServerPort
	server := &http.Server{Addr: addr}

	// Channel for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Run server in a goroutine
	go func() {
		slog.Info("HTTP server listening", "address", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", "error", err)
		}
	}()

	// Wait for interrupt
	<-stop
	slog.Info("Shutdown signal received, cleaning up...")

	// Gracefully shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("HTTP server shutdown error", "error", err)
	}

	// Drop all sandbox databases
	dropSandboxDatabases(cfg)

	slog.Info("Server shutdown complete")
}

func dropSandboxDatabases(cfg *config.Config) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBAdminUser, cfg.DBAdminPassword,
	)
	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		slog.Error("Failed to connect to Postgres", "error", err)
		return
	}
	defer dbConn.Close()

	rows, err := dbConn.Query(`SELECT datname FROM pg_database WHERE datistemplate = false`)
	if err != nil {
		slog.Error("Failed to query databases", "error", err)
		return
	}
	defer rows.Close()

	var dbsToDrop []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			slog.Error("Failed to scan database name", "error", err)
			continue
		}
		if strings.HasPrefix(dbName, "sandbox_") {
			dbsToDrop = append(dbsToDrop, dbName)
		}
	}

	for _, dbName := range dbsToDrop {
		slog.Info("Dropping database", "dbName", dbName)
		if _, err := dbConn.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)); err != nil {
			slog.Error("Failed to drop database", "dbName", dbName, "error", err)
		}
	}
}
