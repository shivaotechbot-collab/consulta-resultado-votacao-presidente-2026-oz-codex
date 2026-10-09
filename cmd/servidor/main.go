package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"consulta-resultado-votacao-presidente-2026-oz-codex/internal/api"
	"consulta-resultado-votacao-presidente-2026-oz-codex/internal/consulta"
)

func main() {
	csv := os.Getenv("CSV_PATH")
	if csv == "" {
		csv = "/home/hermes/projetos/referencia-dados-presidente/osasco-presidente.csv"
	}
	addr := os.Getenv("API_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8096"
	}
	banco, err := consulta.Abrir(csv)
	if err != nil {
		log.Fatalf("importação do CSV: %v", err)
	}
	defer banco.Fechar()
	meta, err := banco.Metadados(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	servidor := &http.Server{Addr: addr, Handler: api.Novo(banco), ReadHeaderTimeout: 5 * time.Second}
	ctx, parar := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer parar()
	go func() {
		<-ctx.Done()
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = servidor.Shutdown(c)
	}()
	log.Printf("API em %s: %d registros, %d seções", addr, meta.Registros, meta.Secoes)
	if err := servidor.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
