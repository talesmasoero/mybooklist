# MyBookList — Contexto para Claude Code

## Sobre o projeto

MyBookList é uma aplicação web para acompanhamento da jornada pessoal de leitura. Foco em registro estruturado de reflexões vinculadas ao ato de leitura e apoio à construção do hábito literário. Projeto acadêmico (Projeto Integrador II — CEUB), em desenvolvimento solo, com IA como ferramenta principal de implementação.

## Stack

- **Backend**: Go (1.26+) com `net/http` da stdlib + roteador Chi
- **Frontend**: React + Vite + React Router + Tailwind CSS + Recharts + Axios, em TypeScript
- **Banco**: PostgreSQL 16
- **Migrations**: golang-migrate
- **Logs**: `log/slog` (stdlib Go), formato JSON
- **Orquestração local**: Docker Compose
- **Integração externa**: Google Books API
- **CI/CD**: GitHub Actions (a configurar)

## Idioma

- **Documentação do projeto** (relatórios da faculdade, README, ADRs): português do Brasil.
- **Código fonte** (variáveis, funções, tipos, comentários, schema do banco, mensagens de log, mensagens de commit): inglês.
- **Mensagens de UI** vistas pelo usuário final: português do Brasil.
- **Conversa interativa** com o desenvolvedor neste terminal: português do Brasil.

## Estrutura do repositório (a criar)

mybooklist/
├── backend/          # API Go
├── frontend/         # SPA React
├── docs/             # Documentação técnica (ARCHITECTURE.md, DATABASE.md, etc.)
├── .claude/          # Configuração do Claude Code (skills, settings)
├── .github/          # Workflows e templates do GitHub
├── docker-compose.yml
├── Makefile          # Será adaptado para Windows: ver scripts/ se necessário
├── .env.example
└── CLAUDE.md         # Este arquivo

## Convenções gerais

- **Commits** seguem Conventional Commits em inglês: `feat:`, `fix:`, `chore:`, `refactor:`, `test:`, `docs:`. Ver skill `write-commit` quando criada.
- **Branches**: `main` é a branch principal. Features em branches `feat/nome-da-feature`. Sem branch de desenvolvimento intermediária (projeto solo).
- **TDD é a abordagem padrão**. Para qualquer feature de lógica: testes primeiro, implementação depois. Ver skill `tdd-cycle` quando criada.
- **Português brasileiro nas conversas comigo**, sempre.

## Glossário PT → EN (mapeamento entidade → tabela/struct)

| Conceitual (docs) | Código/Banco (en) |
|---|---|
| Usuário | User / users |
| Livro | Book / books |
| Leitura | Reading / readings |
| Sessão | Session / sessions |
| Anotação | Note / notes |
| Resenha | Review / reviews |
| Meta | Goal / goals |

## Documentos de referência

(Estes serão criados em `docs/` e referenciados aqui depois.)

- `docs/ARCHITECTURE.md` — Arquitetura do sistema, camadas, deploy.
- `docs/DATABASE.md` — Schema, decisões de modelagem, migrations.
- `docs/API.md` — Convenções REST, formato de erros, autenticação.
- `docs/DESIGN.md` — Identidade visual, paleta, componentes UI.
- `docs/DECISIONS/` — Architecture Decision Records (ADRs).

## Como verificar mudanças

(A preencher depois que houver código rodando.)

## Comportamento esperado de você (Claude)

- Sempre responda em português brasileiro durante o desenvolvimento.
- Quando criar arquivos de código, use inglês.
- Antes de implementar uma feature de lógica, escreva os testes primeiro e mostre-os para mim.
- Antes de criar arquivos novos, verifique se a estrutura de pastas existe; se não, crie.
- Ao final de cada tarefa significativa, me avise para revisar e fazer commit (não commit automático).
- Se eu pedir algo ambíguo, faça uma pergunta antes de assumir.
- Se você ler algo neste arquivo que não bate com a realidade do código (ex.: porque o código mudou), me avise para atualizarmos.