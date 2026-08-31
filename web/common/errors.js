var API_ERROR_MESSAGES = {
  // create / generate
  'dto contains one or more missing fields':                                       'Preencha todos os campos obrigatórios.',
  'dto contains errors: invalid game type':                                        'Tipo de jogo inválido.',
  'dto contains errors: number of games should be greater than zero':              'A quantidade de jogos deve ser maior que zero.',
  'dto contains errors: amount of picked numbers should be within a valid range':  'A quantidade de números por jogo está fora do intervalo permitido para este tipo de jogo.',
  'dto contains errors: amount of fixed numbers cannot be greater than picked numbers': 'A quantidade de números fixos não pode ser maior que a quantidade de números por jogo.',
  'dto contains errors: amount of most sorted numbers cannot be greater than remaining numbers': 'A quantidade de números mais sorteados não pode ser maior que os números restantes.',
  'dto contains errors: a fixed number cannot be a most sorted at same time or vice-versa': 'Um número não pode ser fixo e mais sorteado ao mesmo tempo.',
  'dto contains errors: some fixed numbers are invalid -- choose numbers within a valid range': 'Alguns números fixos estão fora do intervalo válido para este tipo de jogo.',
  'dto contains errors: some most sorted numbers are invalid -- choose numbers within a valid range': 'Alguns números mais sorteados estão fora do intervalo válido para este tipo de jogo.',
  'dto contains errors: number of games is invalid -- use another value or change the amount of fixed numbers': 'Quantidade de jogos inválida. Tente outro valor ou ajuste os números fixos.',
  // conflict
  'combination already exists':          'Esta combinação já foi gerada anteriormente.',
  'combination generated already':       'Esta combinação já foi gerada anteriormente.',
  // not found
  'combination does not exist':          'Combinação não encontrada.',
  'no combination registered with this id': 'Combinação não encontrada.',
  // list
  'no such game type registered':        'Tipo de jogo inválido.',
  // import
  'invalid multipart form':              'Erro ao processar o arquivo enviado.',
  'invalid or missing game_type':        'Tipo de jogo inválido ou não informado.',
  'missing file field':                  'Nenhum arquivo foi enviado.',
  // generate-from
  'invalid json body':                   'Dados inválidos enviados ao servidor.',
  // server
  'internal server error':               'Erro interno do servidor. Tente novamente.'
};

function friendlyError(xhr, fallback) {
  var raw = '';
  try {
    var parsed = JSON.parse(xhr.responseText);
    raw = (parsed.message || xhr.responseText || '').trim().toLowerCase();
  } catch (e) {
    raw = (xhr.responseText || '').trim().toLowerCase();
  }
  return API_ERROR_MESSAGES[raw] || fallback || 'Erro inesperado. Tente novamente.';
}
