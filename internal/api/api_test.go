package api

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"consulta-resultado-votacao-presidente-2026-oz-codex/internal/consulta"
)

func TestAPIConsultasEValidacao(t *testing.T) {
	csv := "aa_eleicao;nm_municipio;nr_turno;ds_cargo;nr_zona;nr_secao;nr_local_votacao;nm_local_votacao;ds_local_votacao_endereco;nr_votavel;nm_votavel;qt_aptos;qt_comparecimento;qt_abstencoes;qt_votos_nominais;qt_votos\n" +
		"2026;OSASCO;1;Presidente;315;74;1066;ESCOLA;RUA;13;CANDIDATA;200;180;20;180;180\n"
	p := filepath.Join(t.TempDir(), "a.csv")
	if err := os.WriteFile(p, []byte(csv), 0600); err != nil {
		t.Fatal(err)
	}
	b, err := consulta.Abrir(p)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Fechar()
	h := Novo(b)
	casos := []struct {
		url    string
		status int
		key    string
	}{
		{"/api/saude", 200, "ok"},
		{"/api/meta", 200, "zonas"},
		{"/api/secoes?zona=315", 200, "74"},
		{"/api/resultado?zona=315&secao=74", 200, "CANDIDATA"},
		{"/api/secoes?zona=999", 200, "[]"},
		{"/api/resultado?zona=315&secao=75", 404, "não encontrada"},
		{"/api/resultado?zona=315&secao=abc", 400, "inválidos"},
		{"/api/secoes?zona=-1", 400, "inválidos"},
		{"/api/resultado?zona=315&secao=74&outro=x", 400, "inválidos"},
		{"/nao-existe", 404, ""},
	}
	for _, c := range casos {
		t.Run(c.url, func(t *testing.T) {
			r := httptest.NewRecorder()
			h.ServeHTTP(r, httptest.NewRequest("GET", c.url, nil))
			if r.Code != c.status || !strings.Contains(r.Body.String(), c.key) {
				t.Fatalf("status %d body %s", r.Code, r.Body.String())
			}
			if strings.HasPrefix(c.url, "/api/") && !strings.HasPrefix(r.Header().Get("Content-Type"), "application/json") {
				t.Fatalf("content type: %s", r.Header().Get("Content-Type"))
			}
			if c.status == 200 && strings.HasPrefix(c.url, "/api/") {
				var value any
				if err := json.Unmarshal(r.Body.Bytes(), &value); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
	r := httptest.NewRecorder()
	h.ServeHTTP(r, httptest.NewRequest("POST", "/api/meta", nil))
	if r.Code != 405 {
		t.Fatalf("POST: %d", r.Code)
	}
}
