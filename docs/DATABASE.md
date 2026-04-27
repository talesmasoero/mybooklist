# DATABASE.md

Documentação do schema do banco de dados PostgreSQL do MyBookList.

---

## Como rodar migrations

As migrations são executadas via container Docker usando a imagem `migrate/migrate`. O serviço está configurado no `docker-compose.yml` sob o profile `migrate`, ou seja, não sobe com `docker compose up` padrão.

### Aplicar todas as migrations pendentes (up)

```bash
docker compose --profile migrate run --rm migrate
```

### Desfazer a última migration (down 1)

```bash
docker compose --profile migrate run --rm migrate \
  -path /migrations \
  -database "postgres://mybooklist:mybooklist_dev@postgres:5432/mybooklist?sslmode=disable" \
  down 1
```

### Ver a versão atual aplicada

```bash
docker compose --profile migrate run --rm migrate \
  -path /migrations \
  -database "postgres://mybooklist:mybooklist_dev@postgres:5432/mybooklist?sslmode=disable" \
  version
```

### Forçar versão (recuperação de estado sujo)

Usado apenas quando a migration falhou no meio e o banco ficou em estado `dirty`:

```bash
docker compose --profile migrate run --rm migrate \
  -path /migrations \
  -database "postgres://mybooklist:mybooklist_dev@postgres:5432/mybooklist?sslmode=disable" \
  force <versão>
```

---

## Convenções

- **Nomenclatura**: `NNNNNN_descricao_da_migration.up.sql` / `.down.sql`, onde `NNNNNN` é um inteiro de 6 dígitos com zero à esquerda (ex: `000001`, `000002`).
- **Par obrigatório**: toda migration deve ter arquivo `.up.sql` e `.down.sql` correspondente.
- **Nunca editar migration já aplicada**: uma vez que um arquivo de migration foi aplicado em qualquer ambiente (dev, staging, prod), ele não deve ser alterado. Para corrigir algo, crie uma nova migration.
- **Idempotência no `.down.sql`**: use `DROP TABLE IF EXISTS`, `DROP INDEX IF EXISTS` etc. para evitar erros ao desfazer em ambientes onde a migration pode não ter sido aplicada.
- **Uma responsabilidade por migration**: cada migration deve alterar uma coisa coesa (criar uma tabela, adicionar uma coluna, criar um índice). Evite migrations que façam tudo ao mesmo tempo.

---

## Schema

### Tabela `users`

Armazena os dados de conta dos usuários da aplicação.

| Coluna          | Tipo         | Constraints                        | Descrição                                                   |
|-----------------|--------------|------------------------------------|-------------------------------------------------------------|
| `id`            | `UUID`       | `PRIMARY KEY`, `DEFAULT gen_random_uuid()` | Identificador único da conta                        |
| `email`         | `TEXT`       | `NOT NULL`, `UNIQUE`               | Endereço de e-mail usado para login                         |
| `password_hash` | `TEXT`       | `NOT NULL`                         | Hash da senha (bcrypt)                                      |
| `name`          | `TEXT`       | `NOT NULL`                         | Nome de exibição do usuário                                 |
| `created_at`    | `TIMESTAMPTZ`| `NOT NULL`, `DEFAULT now()`        | Momento de criação do registro                              |
| `updated_at`    | `TIMESTAMPTZ`| `NOT NULL`, `DEFAULT now()`        | Momento da última atualização (atualizado pela aplicação)   |
| `consented_at`  | `TIMESTAMPTZ`| `NOT NULL`                         | Momento em que o usuário aceitou os termos (requisito LGPD) |

**Constraints adicionais:**

- `users_email_lowercase`: `CHECK (email = LOWER(email))` — garante que o e-mail é sempre armazenado em minúsculas, evitando duplicatas por diferença de capitalização (`User@email.com` vs `user@email.com`).

**Extensão requerida:**

- `pgcrypto`: habilitada no `.up.sql` da migration via `CREATE EXTENSION IF NOT EXISTS pgcrypto`. Necessária para a função `gen_random_uuid()` nas versões do PostgreSQL anteriores à 13 (PostgreSQL 16 já inclui `gen_uuid_v4()` nativamente, mas `pgcrypto` garante compatibilidade e é prática comum).
