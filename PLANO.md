# Consulta de votação presidencial — Osasco (2026)

## Objetivo e fonte
Consulta pública aos votos por zona e seção eleitoral de Osasco, 1º turno da eleição presidencial de 2026, a partir do CSV fornecido em `/home/hermes/projetos/referencia-dados-presidente/osasco-presidente.csv` (separador `;`, codificação Latin-1). Os dados são apresentados como estão na fonte; não afirmar certificação ou atualização em tempo real.

## Superfície e fluxo
**Explore**: área de filtros acima dos resultados. Selecionar zona → seção → ver local, endereço, comparecimento, abstenções e distribuição de votos por candidato, brancos e nulos. Lista ordenada por votos, com busca local por nome ou número e indicadores proporcionais. Compartilhamento via parâmetros de URL. Estados de carregamento, erro e consulta vazia; móvel em coluna única. Estética editorial cívica, sóbria: fundo off-white, tinta azul-petróleo, acento verde escuro, números tabulares, tipografia legível, contraste e foco visível. Sem cartões de marketing ou estatísticas inventadas.

## Arquitetura
- Go HTTP JSON em `127.0.0.1:8096`, endpoints `/api/meta`, `/api/secoes?zona=...`, `/api/resultado?zona=...&secao=...`, `/api/saude`.
- SQLite **em memória** criado e preenchido a cada inicialização a partir do CSV externo; arquivo original não entra no repositório. Índices por zona/seção e candidatos. Importação valida cabeçalhos, números, duplicidade e consistência dos totais de comparecimento por seção.
- HTML5, CSS, JavaScript vanilla separados sob `web/`, servidos **diretamente pelo Nginx**; proxy apenas em `/api/`.
- URL dedicada na porta 8082 do servidor existente; backend somente local, sem afetar os vhosts atuais. Configuração de Nginx e systemd documentadas em `deploy/`.
- TDD: teste falhando → implementação mínima → teste verde → regressão, repetido por comportamento. Testes Go com amostra isolada, testes Node para utilitários de UI e verificações HTTP reais.

## Critérios de aceite
1. Importação do CSV real com 13.950 registros e 1.641 seções; nenhuma cópia do CSV no Git.
2. Consulta por zona e seção devolve votos e dados de local coerentes; parâmetros inválidos ou seção inexistente tratados sem erro interno.
3. Filtros e resultados funcionam no navegador, inclusive em telas móveis e teclado.
4. Nginx entrega arquivos estáticos e encaminha somente API; serviço executa, reinicia e responde após carga em memória.
5. `go test ./...`, testes JS, build e requisições ponta a ponta verificados antes de `git push`; repositório remoto homônimo criado e commit confirmado após push.
