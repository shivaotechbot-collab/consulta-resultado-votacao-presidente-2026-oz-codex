package consulta

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"

	_ "github.com/mattn/go-sqlite3"
)

type Banco struct {
	db        *sql.DB
	registros int
}
type Meta struct {
	Ano       int    `json:"ano"`
	Municipio string `json:"municipio"`
	Turno     int    `json:"turno"`
	Registros int    `json:"registros"`
	Secoes    int    `json:"secoes"`
	Zonas     []int  `json:"zonas"`
}
type Candidato struct {
	Numero int    `json:"numero"`
	Nome   string `json:"nome"`
	Votos  int    `json:"votos"`
}
type Secao struct {
	Zona           int         `json:"zona"`
	Numero         int         `json:"secao"`
	Local          string      `json:"local"`
	Endereco       string      `json:"endereco"`
	Aptos          int         `json:"aptos"`
	Comparecimento int         `json:"comparecimento"`
	Abstencoes     int         `json:"abstencoes"`
	Nominais       int         `json:"nominais"`
	Brancos        int         `json:"brancos"`
	Nulos          int         `json:"nulos"`
	Candidatos     []Candidato `json:"candidatos"`
}

var sequencia atomic.Uint64

// A base existe somente enquanto o processo está ativo. Cada inicialização relê o CSV.
func Abrir(caminho string) (*Banco, error) {
	f, err := os.Open(caminho)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:consulta_votacao_%d?mode=memory&cache=shared&_busy_timeout=5000", sequencia.Add(1)))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	b := &Banco{db: db}
	if err := b.importar(f); err != nil {
		db.Close()
		return nil, err
	}
	return b, nil
}
func (b *Banco) Fechar() error { return b.db.Close() }
func decodificarLatin1(r io.Reader) ([]byte, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var sb strings.Builder
	sb.Grow(len(raw) * 2)
	for _, c := range raw {
		sb.WriteRune(rune(c))
	}
	return []byte(sb.String()), nil
}
func latin1Bytes(s string) ([]byte, error) {
	var b []byte
	for _, r := range s {
		if r > 255 {
			return nil, fmt.Errorf("caractere fora de Latin-1: %c", r)
		}
		b = append(b, byte(r))
	}
	return b, nil
}
func (b *Banco) importar(input io.Reader) error {
	data, err := decodificarLatin1(input)
	if err != nil {
		return err
	}
	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.Comma = ';'
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		return err
	}
	cols := make(map[string]int)
	for i, k := range header {
		cols[strings.TrimSpace(k)] = i
	}
	keys := []string{"aa_eleicao", "nm_municipio", "nr_turno", "ds_cargo", "nr_zona", "nr_secao", "nr_local_votacao", "nm_local_votacao", "ds_local_votacao_endereco", "nr_votavel", "nm_votavel", "qt_aptos", "qt_comparecimento", "qt_abstencoes", "qt_votos_nominais", "qt_votos"}
	for _, k := range keys {
		if _, ok := cols[k]; !ok {
			return fmt.Errorf("coluna ausente: %s", k)
		}
	}
	tx, err := b.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	schema := []string{
		`CREATE TABLE secao (zona INTEGER NOT NULL, numero INTEGER NOT NULL, local TEXT NOT NULL, endereco TEXT NOT NULL, aptos INTEGER NOT NULL, comparecimento INTEGER NOT NULL, abstencoes INTEGER NOT NULL, nominais INTEGER NOT NULL, PRIMARY KEY(zona,numero))`,
		`CREATE TABLE voto (zona INTEGER NOT NULL, secao INTEGER NOT NULL, numero INTEGER NOT NULL, nome TEXT NOT NULL, votos INTEGER NOT NULL, PRIMARY KEY(zona,secao,numero))`,
		`CREATE INDEX idx_voto_secao ON voto(zona,secao)`,
	}
	for _, q := range schema {
		if _, err := tx.Exec(q); err != nil {
			return err
		}
	}
	sec, err := tx.Prepare(`INSERT OR IGNORE INTO secao VALUES (?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer sec.Close()
	voto, err := tx.Prepare(`INSERT INTO voto VALUES (?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer voto.Close()
	n := 0
	for {
		row, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("linha %d: %w", n+2, err)
		}
		get := func(k string) string {
			i := cols[k]
			if i >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[i])
		}
		integer := func(k string) (int, error) {
			v, e := strconv.Atoi(get(k))
			if e != nil || v < 0 {
				return 0, fmt.Errorf("linha %d: %s inválido", n+2, k)
			}
			return v, nil
		}
		if get("aa_eleicao") != "2026" || get("nm_municipio") != "OSASCO" || get("nr_turno") != "1" || get("ds_cargo") != "Presidente" {
			return fmt.Errorf("linha %d: eleição incompatível", n+2)
		}
		vals := make(map[string]int)
		for _, k := range []string{"nr_zona", "nr_secao", "nr_votavel", "qt_aptos", "qt_comparecimento", "qt_abstencoes", "qt_votos_nominais", "qt_votos"} {
			vals[k], err = integer(k)
			if err != nil {
				return err
			}
		}
		z, s := vals["nr_zona"], vals["nr_secao"]
		if vals["qt_aptos"] != vals["qt_comparecimento"]+vals["qt_abstencoes"] {
			return fmt.Errorf("linha %d: aptos inconsistentes", n+2)
		}
		res, err := sec.Exec(z, s, get("nm_local_votacao"), get("ds_local_votacao_endereco"), vals["qt_aptos"], vals["qt_comparecimento"], vals["qt_abstencoes"], vals["qt_votos_nominais"])
		if err != nil {
			return err
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			var local, endereco string
			var aptos, comp, abst, nom int
			err = tx.QueryRow(`SELECT local,endereco,aptos,comparecimento,abstencoes,nominais FROM secao WHERE zona=? AND numero=?`, z, s).Scan(&local, &endereco, &aptos, &comp, &abst, &nom)
			if err != nil {
				return err
			}
			if local != get("nm_local_votacao") || endereco != get("ds_local_votacao_endereco") || aptos != vals["qt_aptos"] || comp != vals["qt_comparecimento"] || abst != vals["qt_abstencoes"] || nom != vals["qt_votos_nominais"] {
				return fmt.Errorf("linha %d: seção inconsistente", n+2)
			}
		}
		if _, err := voto.Exec(z, s, vals["nr_votavel"], get("nm_votavel"), vals["qt_votos"]); err != nil {
			return fmt.Errorf("linha %d: voto duplicado: %w", n+2, err)
		}
		n++
	}
	if n == 0 {
		return errors.New("CSV vazio")
	}
	var invalid int
	err = tx.QueryRow(`SELECT COUNT(*) FROM secao s WHERE
 (SELECT COALESCE(SUM(votos),0) FROM voto v WHERE v.zona=s.zona AND v.secao=s.numero)!=s.comparecimento OR
 (SELECT COALESCE(SUM(votos),0) FROM voto v WHERE v.zona=s.zona AND v.secao=s.numero AND v.numero NOT IN (95,96))!=s.nominais`).Scan(&invalid)
	if err != nil {
		return err
	}
	if invalid > 0 {
		return fmt.Errorf("%d seções com votos inconsistentes", invalid)
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	b.registros = n
	return nil
}
func (b *Banco) Metadados(ctx context.Context) (Meta, error) {
	m := Meta{Ano: 2026, Municipio: "Osasco", Turno: 1, Registros: b.registros, Zonas: []int{}}
	if err := b.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM secao`).Scan(&m.Secoes); err != nil {
		return m, err
	}
	rs, err := b.db.QueryContext(ctx, `SELECT DISTINCT zona FROM secao ORDER BY zona`)
	if err != nil {
		return m, err
	}
	defer rs.Close()
	for rs.Next() {
		var z int
		if err := rs.Scan(&z); err != nil {
			return m, err
		}
		m.Zonas = append(m.Zonas, z)
	}
	return m, rs.Err()
}
func (b *Banco) Secoes(ctx context.Context, z int) ([]int, error) {
	result := []int{}
	rs, err := b.db.QueryContext(ctx, `SELECT numero FROM secao WHERE zona=? ORDER BY numero`, z)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	for rs.Next() {
		var n int
		if err := rs.Scan(&n); err != nil {
			return nil, err
		}
		result = append(result, n)
	}
	return result, rs.Err()
}
func (b *Banco) Resultado(ctx context.Context, z, s int) (*Secao, error) {
	result := &Secao{Zona: z, Numero: s, Candidatos: []Candidato{}}
	err := b.db.QueryRowContext(ctx, `SELECT local,endereco,aptos,comparecimento,abstencoes,nominais FROM secao WHERE zona=? AND numero=?`, z, s).Scan(&result.Local, &result.Endereco, &result.Aptos, &result.Comparecimento, &result.Abstencoes, &result.Nominais)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	rs, err := b.db.QueryContext(ctx, `SELECT numero,nome,votos FROM voto WHERE zona=? AND secao=? ORDER BY votos DESC, numero ASC`, z, s)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	for rs.Next() {
		var c Candidato
		if err := rs.Scan(&c.Numero, &c.Nome, &c.Votos); err != nil {
			return nil, err
		}
		switch c.Numero {
		case 95:
			result.Brancos = c.Votos
		case 96:
			result.Nulos = c.Votos
		default:
			result.Candidatos = append(result.Candidatos, c)
		}
	}
	if err := rs.Err(); err != nil {
		return nil, err
	}
	sort.SliceStable(result.Candidatos, func(i, j int) bool {
		return result.Candidatos[i].Votos > result.Candidatos[j].Votos || (result.Candidatos[i].Votos == result.Candidatos[j].Votos && result.Candidatos[i].Numero < result.Candidatos[j].Numero)
	})
	return result, nil
}
