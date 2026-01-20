package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pouyatavakoli/QueryLab/config"
	"github.com/pouyatavakoli/QueryLab/db"
	"github.com/pouyatavakoli/QueryLab/handler"
)

func main() {
	cfg := config.LoadConfig()

	dbCfg := &db.DBConfig{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		BaseDB:   cfg.DBName,
		InitSQL:  cfg.InitSQL,
	}
	sandbox := db.NewSandboxManager(dbCfg)
	h := handler.NewHandler(sandbox)

	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/api/session", h.CreateSession)
	http.HandleFunc("/api/query", h.RunQuery)

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Println("QueryLab server running on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
