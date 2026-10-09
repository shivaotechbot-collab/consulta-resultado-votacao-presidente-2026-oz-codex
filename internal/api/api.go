package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"consulta-resultado-votacao-presidente-2026-oz-codex/internal/consulta"
)

func enviar(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
func erro(w http.ResponseWriter, code int, msg string) {
	enviar(w, code, map[string]string{"erro": msg})
}
func inteiro(r *http.Request, nome string) (int, bool) {
	v := r.URL.Query().Get(nome)
	if v == "" || strings.Trim(v, "0123456789") != "" {
		return 0, false
	}
	n, e := strconv.Atoi(v)
	return n, e == nil && n > 0
}
func Novo(b *consulta.Banco) http.Handler {
	mux := http.NewServeMux()
	somenteGet := func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method != "GET" {
			w.Header().Set("Allow", "GET")
			erro(w, 405, "método não permitido")
			return false
		}
		return true
	}
	parametros := func(w http.ResponseWriter, r *http.Request, nomes ...string) bool {
		allowed := make(map[string]bool)
		for _, n := range nomes {
			allowed[n] = true
		}
		for k, v := range r.URL.Query() {
			if !allowed[k] || len(v) != 1 {
				erro(w, 400, "parâmetros inválidos")
				return false
			}
		}
		return true
	}
	mux.HandleFunc("/api/saude", func(w http.ResponseWriter, r *http.Request) {
		if !somenteGet(w, r) {
			return
		}
		enviar(w, 200, map[string]string{"estado": "ok"})
	})
	mux.HandleFunc("/api/meta", func(w http.ResponseWriter, r *http.Request) {
		if !somenteGet(w, r) {
			return
		}
		m, e := b.Metadados(r.Context())
		if e != nil {
			erro(w, 500, "falha na consulta")
			return
		}
		enviar(w, 200, m)
	})
	mux.HandleFunc("/api/secoes", func(w http.ResponseWriter, r *http.Request) {
		if !somenteGet(w, r) || !parametros(w, r, "zona") {
			return
		}
		z, ok := inteiro(r, "zona")
		if !ok {
			erro(w, 400, "parâmetros inválidos")
			return
		}
		secoes, e := b.Secoes(r.Context(), z)
		if e != nil {
			erro(w, 500, "falha na consulta")
			return
		}
		enviar(w, 200, secoes)
	})
	mux.HandleFunc("/api/resultado", func(w http.ResponseWriter, r *http.Request) {
		if !somenteGet(w, r) || !parametros(w, r, "zona", "secao") {
			return
		}
		z, okZ := inteiro(r, "zona")
		s, okS := inteiro(r, "secao")
		if !okZ || !okS {
			erro(w, 400, "parâmetros inválidos")
			return
		}
		res, e := b.Resultado(r.Context(), z, s)
		if e != nil {
			erro(w, 500, "falha na consulta")
			return
		}
		if res == nil {
			erro(w, 404, "seção não encontrada")
			return
		}
		enviar(w, 200, res)
	})
	return mux
}
