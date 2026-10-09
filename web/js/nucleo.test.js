const test = require('node:test');
const assert = require('node:assert/strict');
const { formatoInteiro, percentual, filtrarCandidatos, lerConsulta } = require('./nucleo.js');

test('formata milhares em português brasileiro', () => {
  assert.equal(formatoInteiro(1641), '1.641');
  assert.equal(formatoInteiro(0), '0');
});
test('percentual evita divisão por zero e usa uma casa decimal', () => {
  assert.equal(percentual(82, 161), '50,9%');
  assert.equal(percentual(12, 0), '0,0%');
});
test('filtra candidatos por nome acentuado ou número sem alterar a lista original', () => {
  const dados = [{ numero: 13, nome: 'LUIZ INÁCIO LULA DA SILVA', votos: 62 }, { numero: 22, nome: 'FLAVIO NANTES BOLSONARO', votos: 82 }];
  assert.deepEqual(filtrarCandidatos(dados, 'inacio').map(x => x.numero), [13]);
  assert.deepEqual(filtrarCandidatos(dados, '22').map(x => x.numero), [22]);
  assert.equal(dados.length, 2);
});
test('lê somente zona e seção positivas da URL', () => {
  assert.deepEqual(lerConsulta('?zona=315&secao=74'), { zona: 315, secao: 74 });
  assert.deepEqual(lerConsulta('?zona=-1&secao=x'), { zona: null, secao: null });
  assert.deepEqual(lerConsulta('?zona=315&secao='), { zona: 315, secao: null });
});
