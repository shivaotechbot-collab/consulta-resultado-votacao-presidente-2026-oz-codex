package consulta

import (
	"context"
	"testing"
)

func TestBancosEmMemoriaSaoIsolados(t *testing.T) {
	a := carregarTeste(t, cabecalho+linhas)
	b := carregarTeste(t, cabecalho+linhas)
	ma, err := a.Metadados(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	mb, err := b.Metadados(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ma.Registros != 5 || mb.Registros != 5 || ma.Secoes != 2 || mb.Secoes != 2 {
		t.Fatalf("bancos: %+v %+v", ma, mb)
	}
}
