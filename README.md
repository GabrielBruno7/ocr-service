# OCR Service

Serviço para extração e classificação automática de dados de documentos de identidade (CNH e RG) a partir de imagens, usando OCR.

## O que o projeto faz

1. O cliente envia uma imagem de um documento via API HTTP.
2. O documento é salvo e uma mensagem é publicada numa fila (RabbitMQ) para processamento assíncrono.
3. Um worker consome a fila, roda OCR (Tesseract) sobre a imagem, classifica o tipo do documento (CNH, RG ou desconhecido) e extrai campos estruturados (nome, CPF).
4. O resultado é persistido no banco (Postgres) e pode ser consultado depois via API.

## Arquitetura

O projeto segue os princípios de Clean Architecture, organizado por camadas dentro de cada módulo de domínio

## Stack

- **Linguagem:** Go
- **HTTP:** [Gin](https://github.com/gin-gonic/gin)
- **Banco de dados:** PostgreSQL (via [pgx](https://github.com/jackc/pgx))
- **Fila de mensagens:** RabbitMQ
- **OCR:** Tesseract
- **Testes:** [testify](https://github.com/stretchr/testify)

## Como rodar

### Pré-requisitos

- Docker e Docker Compose

### Subindo o ambiente

```bash
docker compose up --build
```

Isso sobe:
- `api` — servidor HTTP
- `worker` — consumidor da fila de OCR
- `rabbitmq` — fila de mensagens
- `postgres` — banco de dados

### Variáveis de ambiente

Copie o `.env.example` para `.env` e ajuste conforme necessário:

```bash
cp .env.example .env
```

| Variável | Descrição |
|---|---|
| `DATABASE_URL` | String de conexão com o Postgres |
| `AMQP_URL` | String de conexão com o RabbitMQ |
| `UPLOAD_DIR` | Diretório onde os arquivos enviados são salvos |
| `PORT` | Porta em que a API roda |
| `OCR_LANGUAGE` | Idioma usado pelo Tesseract |
| `OCR_POOL_SIZE` | Tamanho do pool de processadores OCR |
| `PROCESSING_TIMEOUT` | Tempo limite antes de reprocessar um documento travado |

*(ajuste a tabela acima conforme as variáveis reais do seu `config.go`)*

## Endpoints

### `POST /documents/extract`

Envia um documento (multipart/form-data, campo `file`) para processamento.

**Resposta (202 Accepted):**
```json
{
  "id": "e6a1...",
  "status": "pending"
}
```

### `GET /documents/:id`

Consulta o status e resultado do processamento de um documento.

**Resposta (200 OK):**
```json
{
  "id": "e6a1...",
  "status": "done",
  "file_name": "cnh.png",
  "document_type": "cnh",
  "extracted_text": "...",
  "extracted_fields": {
    "name": "JOAO DA SILVA",
    "cpf": "123.456.789-00"
  },
  "error_message": null
}
```

Status possíveis: `pending`, `processing`, `done`, `failed`.

## Rodando os testes

```bash
go test ./...
```

Os testes de handler e usecase usam um repositório em memória (`domain.MemoryRepository`), sem necessidade de banco de dados real.

## Migrations

```bash
[comando de migration usado no projeto, ex: migrate -path migrations -database $DATABASE_URL up]
```
