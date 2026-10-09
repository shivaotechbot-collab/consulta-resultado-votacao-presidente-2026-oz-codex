# Consulta de resultado da votação presidencial 2026 — Osasco

Frontend HTML5/CSS/JavaScript vanilla servido diretamente pelo Nginx; API Go que importa o CSV Latin-1 para SQLite **em memória** a cada inicialização. Sem escrita em disco do banco. Uso exploratório; não é ferramenta oficial da Justiça Eleitoral.

## Executar

Pré-requisitos: Go 1.22+, compilador C (driver `go-sqlite3`), Node.js para os testes de UI, Nginx para servir o frontend.

```sh
export CSV_PATH=/caminho/para/osasco-presidente.csv
export API_ADDR=127.0.0.1:8096
CGO_ENABLED=1 go run ./cmd/servidor
```

Dados esperados: CSV separado por ponto e vírgula, em ISO-8859-1, 1º turno, cargo Presidente, município OSASCO. A fonte de dados não é versionada.

API: `GET /api/saude`, `GET /api/meta`, `GET /api/secoes?zona=315`, `GET /api/resultado?zona=315&secao=74`. Respostas JSON, erros 400/404/405. O frontend permite filtrar candidatos localmente, preserva a consulta na URL e exibe votos nominais, brancos e nulos separadamente. Percentuais dos candidatos usam como denominador os votos nominais.

## Testes

```sh
CSV_TEST_PATH=/caminho/para/osasco-presidente.csv go test ./... -count=1
node --test web/js/*.test.js
go vet ./...
go build -o bin/servidor ./cmd/servidor
```

Os testes unitários utilizam CSVs pequenos; o teste de integração com a fonte completa roda quando `CSV_TEST_PATH` é informado. Testes seguem ciclos RED → GREEN → regressão.

## Publicação neste servidor

`deploy/nginx.conf` usa a porta HTTP **8082**, independente das aplicações anteriores em 8080 e 8081; `deploy/consulta-votacao-codex.service` executa a API somente em loopback **8096**. Ajuste caminhos caso publique em outra máquina. Copie `web/` para `/var/www/consulta-resultado-votacao-presidente-2026-oz-codex`, instale a configuração Nginx em `sites-available` e a unit no systemd; valide `nginx -t` e reinicie/recarregue. URL no servidor: `http://200.234.212.203:8082/`. A porta pública precisa estar liberada no firewall do provedor para acesso externo.

O `web/js/nucleo.test.js` é excluído pelo Nginx. O CSV e o binário não são enviados ao GitHub. A aplicação mostra apenas os dados do arquivo fornecido, sem garantir autenticidade oficial nem atualização em tempo real.
