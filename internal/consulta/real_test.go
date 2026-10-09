package consulta

import (
	"context"
	"os"
	"testing"
)

func TestCSVRealQuandoDisponivel(t *testing.T) {
	path := os.Getenv("CSV_TEST_PATH")
	if path == "" {
		t.Skip("defina CSV_TEST_PATH para teste com a fonte real")
	}
	b, err := Abrir(path)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Fechar()
	meta, err := b.Metadados(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if meta.Registros != 13950 || meta.Secoes != 1641 || len(meta.Zonas) != 6 {
		t.Fatalf("contagem inesperada: %+v", meta)
	}
	got, err := b.Resultado(context.Background(), 315, 74)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.Comparecimento != 174 || got.Nominais != 161 || got.Brancos != 4 || got.Nulos != 9 || got.Local != "EMEF, OSVALDO QUIRINO SIMÕES" || got.Candidatos[0].Numero != 22 || got.Candidatos[0].Votos != 82 {
		t.Fatalf("resultado inesperado: %+v", got)
	}
}
