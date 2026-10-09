package consulta

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const cabecalho = "aa_eleicao;nm_municipio;nr_turno;ds_cargo;nr_zona;nr_secao;nr_local_votacao;nm_local_votacao;ds_local_votacao_endereco;nr_votavel;nm_votavel;qt_aptos;qt_comparecimento;qt_abstencoes;qt_votos_nominais;qt_votos\n"
const linhas = "2026;OSASCO;1;Presidente;315;74;1066;ESCOLA SÃO JOSÉ;RUA 1;13;CANDIDATA A;200;180;20;173;100\n" +
	"2026;OSASCO;1;Presidente;315;74;1066;ESCOLA SÃO JOSÉ;RUA 1;22;CANDIDATO B;200;180;20;173;73\n" +
	"2026;OSASCO;1;Presidente;315;74;1066;ESCOLA SÃO JOSÉ;RUA 1;95;VOTO BRANCO;200;180;20;173;4\n" +
	"2026;OSASCO;1;Presidente;315;74;1066;ESCOLA SÃO JOSÉ;RUA 1;96;VOTO NULO;200;180;20;173;3\n" +
	"2026;OSASCO;1;Presidente;277;8;1200;OUTRA ESCOLA;AV. 2;13;CANDIDATA A;100;90;10;90;90\n"

func carregarTeste(t *testing.T, conteudo string) *Banco {
	t.Helper()
	p := filepath.Join(t.TempDir(), "amostra.csv")
	if err := os.WriteFile(p, []byte(conteudo), 0600); err != nil {
		t.Fatal(err)
	}
	b, err := Abrir(p)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { b.Fechar() })
	return b
}

func TestImportaCSVLatin1EConsultaSecao(t *testing.T) {
	amostra, err := latin1Bytes(cabecalho + linhas)
	if err != nil {
		t.Fatal(err)
	}
	b := carregarTeste(t, string(amostra))
	meta, err := b.Metadados(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if meta.Registros != 5 || meta.Secoes != 2 || len(meta.Zonas) != 2 || meta.Zonas[0] != 277 || meta.Zonas[1] != 315 {
		t.Fatalf("metadados: %+v", meta)
	}
	secoes, err := b.Secoes(context.Background(), 315)
	if err != nil || len(secoes) != 1 || secoes[0] != 74 {
		t.Fatalf("secoes: %v %v", secoes, err)
	}
	resultado, err := b.Resultado(context.Background(), 315, 74)
	if err != nil {
		t.Fatal(err)
	}
	if resultado == nil || resultado.Local != "ESCOLA SÃO JOSÉ" || resultado.Comparecimento != 180 || resultado.Abstencoes != 20 || resultado.Aptos != 200 || resultado.Nominais != 173 || resultado.Brancos != 4 || resultado.Nulos != 3 || len(resultado.Candidatos) != 2 || resultado.Candidatos[0].Numero != 13 || resultado.Candidatos[0].Votos != 100 || resultado.Candidatos[1].Votos != 73 {
		t.Fatalf("resultado: %+v", resultado)
	}
}

func TestRejeitaCSVInconsistente(t *testing.T) {
	casos := map[string]string{
		"sem colunas":               "nr_zona;nr_secao\n315;74\n",
		"votos negativos":           cabecalho + strings.Replace(linhas, ";100\n", ";-1\n", 1),
		"registro duplicado":        cabecalho + linhas + strings.Split(linhas, "\n")[0] + "\n",
		"comparecimento divergente": cabecalho + strings.Replace(linhas, ";180;20;173;73", ";181;20;173;73", 1),
	}
	for nome, conteudo := range casos {
		t.Run(nome, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "errado.csv")
			if err := os.WriteFile(p, []byte(conteudo), 0600); err != nil {
				t.Fatal(err)
			}
			b, err := Abrir(p)
			if err == nil {
				b.Fechar()
				t.Fatal("CSV inválido aceito")
			}
		})
	}
}

func TestSecaoInexistente(t *testing.T) {
	b := carregarTeste(t, cabecalho+linhas)
	got, err := b.Resultado(context.Background(), 315, 999)
	if err != nil || got != nil {
		t.Fatalf("got=%v err=%v", got, err)
	}
}
