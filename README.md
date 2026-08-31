# Combina v3.0

Um projeto em parceria com meu pai para gerar e gerenciar jogos de loteria brasileira. As versões anteriores (v1 em Java, v2 em JS puro) não foram versionadas. A v3 introduz uma API REST, persistência em PostgreSQL e uma interface web.

## O que o projeto faz

- **Gera combinações** de jogos para Lotofácil, Mega-Sena, Quina, Lotomania, Quina Brasil, Quininha e Seninha
- **Importa jogos** a partir de um arquivo de texto (um jogo por linha, números separados por vírgula), com validação automática por tipo de jogo
- **Gera jogos adicionais** a partir de um conjunto já existente, garantindo que os novos jogos não repitam nenhum dos jogos importados nem entre si
- **Persiste e gerencia** as combinações geradas via API REST (criar, listar, buscar, deletar, conferir resultado)
- Suporta **números fixos** (presentes em todo jogo) e **números mais sorteados** (com maior probabilidade de aparecer) na geração

## Pré-requisitos

- Docker e Docker Compose (recomendado), **ou**
- Go 1.24+ e PostgreSQL rodando em `localhost:5432`

## Executar com Docker (recomendado)

```bash
# Primeira vez: inicializa o banco de dados
docker compose up -d postgres
docker compose run --rm app ./combina -init-db

# Nas demais vezes: sobe tudo com um comando
docker compose up
```

Acesse a interface em `http://localhost:3000/combinar.html`.

## Executar localmente (sem Docker)

1. Clone o repositório e entre na pasta:
   ```bash
   git clone <url-do-repositório>
   cd combina
   ```

2. Crie o banco de dados (apenas na primeira vez):
   ```bash
   go run src/main.go -init-db
   ```

3. Inicie o servidor:
   ```bash
   go run src/main.go
   ```

O servidor sobe na porta `3000`. A interface estará em `http://localhost:3000/combinar.html`.



Base URL: `http://localhost:3000`

### Listar combinações
```
GET /combinations?type={tipo}
```
Parâmetro `type` opcional. Tipos válidos: `Lotofacil`, `Mega-Sena`, `Quina`, `Lotomania`, `Quina-Brasil`, `Quininha`, `Seninha`.

### Criar combinação
```
POST /combinations
Content-Type: application/json
```
```json
{
  "game_type": "Lotofacil",
  "num_games": 10,
  "num_each": 15,
  "fixed_numbers": [3, 7],
  "most_sorted": [1, 5, 12, 20],
  "alias": "meu-jogo"
}
```
`fixed_numbers` e `most_sorted` são opcionais. Retorna `201 Created` com o ID no header `Location`.

### Importar jogos de arquivo
```
POST /combinations/import
Content-Type: multipart/form-data
```
Campos: `game_type` (string) e `file` (arquivo de texto).

Formato do arquivo — cada linha é um jogo, números separados por vírgula:
```
1,4,7,12,15,18,20,22,23,24,25
2,5,8,11,14,17,19,21,23,24,25
```
Linhas inválidas (número fora do intervalo, quantidade incorreta, repetição) são descartadas. Retorna:
```json
{ "games": [[...], ...], "game_type": "Lotofacil", "discarded": 2 }
```

### Gerar jogos a partir de um conjunto existente
```
POST /combinations/generate-from
Content-Type: application/json
```
```json
{
  "game_type": "Lotofacil",
  "num_games": 5,
  "num_each": 15,
  "existing_games": [[1,4,7,...], [2,5,8,...]],
  "alias": "gerados"
}
```
Os jogos gerados não duplicam nenhum jogo em `existing_games`. Retorna `201 Created` com o objeto `Lotto` completo.

### Buscar combinação
```
GET /combinations/{id}
```

### Deletar combinação
```
DELETE /combinations/{id}
```

### Conferir resultado
```
GET /combinations/evaluate/{id}?values=[15,20,22,...]
```
Retorna um mapa com a contagem de acertos por jogo, filtrado apenas pelos prêmios válidos do tipo de jogo.

## Testes

```bash
go test ./...
```

## Regras de validação por tipo de jogo

| Tipo | Intervalo | Números por jogo |
|---|---|---|
| Lotofácil | 1–25 | 15 a 18 |
| Mega-Sena | 1–60 | 6 a 15 |
| Quina | 1–80 | 5 a 15 |
| Lotomania | 0–99 | 50 |
| Quina Brasil | 1–80 | 13 |
| Quininha | 1–80 | 15 a 20, 25 ou 30 |
| Seninha | 1–60 | 14 a 20, 25 ou 30 |
