(() => {
  'use strict';
  const { formatoInteiro, percentual, filtrarCandidatos, lerConsulta } = window.ConsultaNucleo;
  const $ = id => document.getElementById(id);
  const zona = $('zona'), secao = $('secao'), estado = $('estado'), resultado = $('resultado');
  let dados = null, versao = 0;
  function mensagem(texto, erro = false) { estado.textContent = texto; estado.classList.toggle('erro', erro); }
  function opcao(select, value, label) { const item = document.createElement('option'); item.value = value; item.textContent = label; select.append(item); }
  async function obter(url) {
    const resposta = await fetch(url, { headers: { Accept: 'application/json' } });
    if (!resposta.ok) throw new Error(resposta.status === 404 ? 'Seção não encontrada.' : 'Não foi possível carregar os dados. Tente novamente.');
    return resposta.json();
  }
  function atualizarLista() {
    const encontrados = filtrarCandidatos(dados.candidatos, $('busca').value);
    const lista = $('candidatos'); lista.replaceChildren();
    $('sem-candidatos').classList.toggle('oculto', encontrados.length > 0);
    encontrados.forEach(c => {
      const li = document.createElement('li'); li.className = 'candidato';
      const numero = document.createElement('span'); numero.className = 'numero'; numero.textContent = c.numero;
      const corpo = document.createElement('div');
      const nome = document.createElement('span'); nome.className = 'nome'; nome.textContent = c.nome;
      const barra = document.createElement('div'); barra.className = 'barra-votos'; barra.setAttribute('aria-hidden', 'true');
      const preenchimento = document.createElement('div'); preenchimento.className = 'preenchimento'; preenchimento.style.width = Math.min(100, dados.nominais ? c.votos * 100 / dados.nominais : 0) + '%';
      barra.append(preenchimento); corpo.append(nome, barra);
      const votos = document.createElement('div'); votos.className = 'votos'; votos.append(document.createTextNode(formatoInteiro(c.votos)));
      const taxa = document.createElement('small'); taxa.textContent = percentual(c.votos, dados.nominais); votos.append(taxa);
      li.append(numero, corpo, votos); lista.append(li);
    });
  }
  function exibir(d) {
    dados = d;
    $('referencia').textContent = `ZONA ${d.zona} / SEÇÃO ${d.secao}`;
    $('local').textContent = d.local; $('endereco').textContent = d.endereco;
    for (const campo of ['aptos', 'comparecimento', 'abstencoes', 'nominais', 'brancos', 'nulos']) $(campo).textContent = formatoInteiro(d[campo]);
    $('busca').value = ''; atualizarLista(); resultado.classList.remove('oculto'); mensagem('');
  }
  async function carregarSecoes(selecao = null) {
    const atual = ++versao; dados = null; resultado.classList.add('oculto');
    secao.replaceChildren(); secao.disabled = true;
    if (!zona.value) { opcao(secao, '', 'Selecione uma zona primeiro'); mensagem('Selecione uma zona e uma seção para começar.'); return; }
    opcao(secao, '', 'Carregando seções…'); mensagem('Carregando seções…');
    try {
      const valores = await obter(`/api/secoes?zona=${encodeURIComponent(zona.value)}`);
      if (atual !== versao) return;
      secao.replaceChildren(); opcao(secao, '', 'Selecione uma seção');
      valores.forEach(n => opcao(secao, n, `Seção ${n}`)); secao.disabled = false;
      if (selecao && valores.includes(selecao)) secao.value = String(selecao);
      mensagem('Selecione a seção e clique em “Ver resultado”.');
    } catch (e) { if (atual === versao) { secao.replaceChildren(); opcao(secao, '', 'Não foi possível carregar'); mensagem(e.message, true); } }
  }
  async function consultar() {
    const atual = ++versao; resultado.classList.add('oculto'); mensagem('Carregando resultado…');
    const z = zona.value, s = secao.value;
    try {
      const d = await obter(`/api/resultado?zona=${encodeURIComponent(z)}&secao=${encodeURIComponent(s)}`);
      if (atual !== versao) return;
      exibir(d); history.replaceState(null, '', `?zona=${encodeURIComponent(z)}&secao=${encodeURIComponent(s)}`);
      resultado.scrollIntoView({ behavior: 'smooth', block: 'start' });
    } catch (e) { if (atual === versao) mensagem(e.message, true); }
  }
  zona.addEventListener('change', () => { history.replaceState(null, '', location.pathname); carregarSecoes(); });
  secao.addEventListener('change', () => { resultado.classList.add('oculto'); mensagem('Clique em “Ver resultado” para consultar esta seção.'); });
  $('form-consulta').addEventListener('submit', e => { e.preventDefault(); if (zona.value && secao.value) consultar(); });
  $('busca').addEventListener('input', () => { if (dados) atualizarLista(); });
  (async () => {
    try {
      const meta = await obter('/api/meta'); zona.replaceChildren(); opcao(zona, '', 'Selecione uma zona'); meta.zonas.forEach(n => opcao(zona, n, `Zona ${n}`));
      const inicial = lerConsulta(location.search);
      if (meta.zonas.includes(inicial.zona)) { zona.value = String(inicial.zona); await carregarSecoes(inicial.secao); if (inicial.secao && secao.value === String(inicial.secao)) await consultar(); }
      else mensagem('Selecione uma zona e uma seção para começar.');
    } catch (e) { zona.replaceChildren(); opcao(zona, '', 'Não foi possível carregar'); mensagem(e.message, true); }
  })();
})();
