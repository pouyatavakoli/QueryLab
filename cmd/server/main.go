package main

import (
	"log"
	"net/http"

	"github.com/pouyatavakoli/QueryLab/config"
	"github.com/pouyatavakoli/QueryLab/db"
	"github.com/pouyatavakoli/QueryLab/handler"
)

func main() {
	cfg := config.LoadConfig()

	sandbox := db.NewSandboxManager(&db.DBConfig{
		Host: cfg.DBHost,
		Port: cfg.DBPort,

		AdminUser:     cfg.DBAdminUser,
		AdminPassword: cfg.DBAdminPassword,

		SandboxUser:     cfg.DBSandboxUser,
		SandboxPassword: cfg.DBSandboxPassword,

		BaseDB:  cfg.DBName,
		InitSQL: cfg.InitSQL,
	})

	h := handler.NewHandler(sandbox)

	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/api/session", h.CreateSession)
	http.HandleFunc("/api/query", h.RunQuery)

	addr := ":" + cfg.ServerPort
	log.Println("QueryLab listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
