# Backend Go — Convenções

## Camadas

Fluxo estrito: `domain` ← `repositories` ← `services` ← `handlers`.
Nunca pule camadas (ex.: handler não acessa repository diretamente).

- `domain/` — structs e interfaces de domínio, sem dependências externas
- `repositories/` — acesso ao banco; recebe `*sql.DB`, retorna tipos de `domain`
- `services/` — lógica de negócio; recebe interfaces de repository
- `handlers/` — HTTP: decodifica request, chama service, escreve response
- `middleware/` — middlewares HTTP reutilizáveis
- `config/` — carregamento de variáveis de ambiente

## Erros

Defina `AppError` em `domain/errors.go`:
```go
type AppError struct {
    Code    int
    Message string
    Err     error
}
```
Handlers traduzem `AppError` para JSON. Nunca use `fmt.Errorf` diretamente nos handlers.

## Testes

Table-driven, sempre com `t.Run`:
```go
tests := []struct {
    name string
    // ...
}{
    {name: "success case", ...},
    {name: "error case", ...},
}
for _, tc := range tests {
    t.Run(tc.name, func(t *testing.T) { ... })
}
```
Arquivos de teste: `foo_test.go` ao lado de `foo.go`, mesmo pacote ou `_test` para testes de caixa preta.

## Imports

Três blocos separados por linha em branco, nesta ordem:
```go
import (
    // stdlib
    "context"
    "net/http"

    // external
    "github.com/go-chi/chi/v5"

    // internal
    "github.com/talesmasoero/mybooklist/backend/internal/domain"
)
```

## Variáveis de ambiente

Em desenvolvimento, o servidor carrega automaticamente o arquivo `.env` da raiz do repositório via `godotenv` (caminho `../.env` relativo ao diretório `backend/`). Se o arquivo não existir, um warning é logado e a execução continua normalmente — o comportamento esperado em produção, onde as variáveis vêm do ambiente do provedor (Railway, Fly.io, etc.) e nenhum `.env` é necessário.

Nunca commitar o `.env` real; o `.env.example` na raiz documenta as variáveis necessárias.

## Logs

Use `slog` com contexto propagado. Nunca use `fmt.Println` ou `log.Printf`.
```go
slog.InfoContext(ctx, "event description", "key", value)
slog.ErrorContext(ctx, "error description", "error", err)
```
Campos obrigatórios em logs de request: `method`, `path`, `status`, `duration_ms`.
