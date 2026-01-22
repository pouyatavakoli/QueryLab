package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/pouyatavakoli/QueryLab/config"
	"github.com/pouyatavakoli/QueryLab/db"
	"github.com/pouyatavakoli/QueryLab/handler"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Log startup with version/context if available
	slog.Info("starting QueryLab server",
		"timestamp", time.Now().Format(time.RFC3339),
		"pid", os.Getpid(),
	)

	// Configuration loading
	slog.Info("loading configuration")
	cfg := config.LoadConfig()
	slog.Info("configuration loaded successfully",
		"server_port", cfg.ServerPort,
		"db_host", cfg.DBHost,
		"db_port", cfg.DBPort,
	)

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
		SessionTimeout: 1 * time.Hour, // Sessions expire after 1 hour of inactivity
	})
	slog.Info("sandbox database manager initialized")

	// Handler initialization
	slog.Info("initializing HTTP handlers")
	h := handler.NewHandler(sandbox)

	// Route setup
	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/api/session", h.CreateSession)
	http.HandleFunc("/api/query", h.RunQuery)
	http.HandleFunc("/api/logout", h.Logout)
	http.HandleFunc("/api/health", h.HealthCheck)

	slog.Info("HTTP routes registered",
		"static_files", "./frontend",
		"api_endpoints", []string{
			"/api/session",
			"/api/query",
			"/api/logout",
			"/api/health",
		},
	)

	// Server startup
	addr := ":" + cfg.ServerPort
	slog.Info("starting HTTP server",
		"address", addr,
		"listen_address", "http://localhost"+addr,
	)

	// Log server shutdown gracefully
	slog.Info("HTTP server listening for requests")
	err := http.ListenAndServe(addr, nil)
	if err != nil && err != http.ErrServerClosed {
		slog.Error("HTTP server failed to start",
			"error", err,
			"address", addr,
		)
		log.Fatal(err)
	}
}