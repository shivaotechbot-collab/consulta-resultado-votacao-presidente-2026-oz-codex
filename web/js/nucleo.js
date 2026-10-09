(function (root) {
  'use strict';
  function formatoInteiro(n) { return new Intl.NumberFormat('pt-BR').format(n); }
  function percentual(parte, total) { return new Intl.NumberFormat('pt-BR', { minimumFractionDigits: 1, maximumFractionDigits: 1 }).format(total ? parte * 100 / total : 0) + '%'; }
  function filtrarCandidatos(dados, busca) {
    const normalizar = s => String(s).normalize('NFD').replace(/[\u0300-\u036f]/g, '').toLowerCase();
    const q = normalizar(busca.trim());
    return dados.filter(item => normalizar(item.nome).includes(q) || String(item.numero).includes(q));
  }
  function lerConsulta(query) {
    const params = new URLSearchParams(query);
    const numero = nome => { const s = params.get(nome); return s && /^[1-9]\d*$/.test(s) && Number.isSafeInteger(Number(s)) ? Number(s) : null; };
    return { zona: numero('zona'), secao: numero('secao') };
  }
  const nucleo = { formatoInteiro, percentual, filtrarCandidatos, lerConsulta };
  if (typeof module !== 'undefined' && module.exports) module.exports = nucleo;
  root.ConsultaNucleo = nucleo;
})(typeof window !== 'undefined' ? window : globalThis);
