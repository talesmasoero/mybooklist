# MyBookList — Project Specification

**This document is the single source of truth for product decisions, requirements, and conceptual modeling for the MyBookList project. All other documentation (`ARCHITECTURE.md`, `API.md`, ADRs) and all generated code must be consistent with this file.**

This is the English translation and reorganization of the academic report `mybooklist_documento.docx`, intended for consumption by Claude Code and any developer joining the project.

---

## 1. Vision

MyBookList is a web application for tracking the personal reading journey, with two complementary goals:

1. **Reading journey** — let users document their reading experience in a structured, introspective way and track their progress over time.
2. **Habit building** — help readers who want to develop the reading habit stay motivated through goals, progress visualization, and metrics.

The product positions itself as a complementary, not competing, tool to platforms like Skoob and Goodreads. Those platforms serve users focused on social discovery and broad cataloging well; MyBookList serves users focused on individual journey and habit development.

Social features (public reviews, visible profiles) exist but are intentionally secondary. The home screen shows the user's current reading progress, not aggregated content from other users.

## 2. Identified Problem

Three related but distinct gaps exist in the current landscape of reader tools:

1. **No active habit support.** Existing solutions offer cataloging and social engagement but treat motivation features (annual goals, progress tracking, pace metrics) as secondary, not core. For aspiring readers without an established habit, this means the platform plays no active role in encouraging continuity.

2. **No structured reflection tied to the act of reading.** Commercial apps offer either (i) reviews that refer to the book as a whole, or (ii) free notes in generic tools (Notion, Google Docs, Word, Apple Notes), disconnected from the reading session. No tool in the surveyed market structures the record "I read pages X to Y on this date and these are my notes about those specific passages." Readers wanting to preserve granular reflections fall back to physical notebooks or scattered text documents.

3. **Outdated UX/UI** in the main literary platforms relative to current standards. Goodreads in particular is widely cited for visually dated interface and mobile-poor flows. This drives away new users (especially aspiring readers, sensitive to first contact) and reduces engagement of regular readers.

## 3. Target Audience

Civilian community interested in literature. Two complementary profiles:

- **Avid readers** with established habits, needing robust organization for history and personal notes. Differentiator: granular session and note structure.
- **Aspiring readers** seeking motivation and goal structure to develop the habit. Differentiator: tracking features (goals, progress bar, evolution dashboard) operating from the first screen.

## 4. Goals

### General

Develop a responsive web application for managing and tracking the personal reading journey, simultaneously focused on (a) structured reflection tied to the act of reading and (b) reading habit building through goals and progress tracking.

### Specific

- Implement secure authentication: registration, login, logout, password recovery via email.
- Integrate Google Books API for automated book search and registration, with local cache and manual fallback.
- Develop **Library** module covering states "Want to Read", "Reading", "Read", "Abandoned", with filters and ordering.
- Create **Reading Journal** module for session logging with automatic progress calculation, optional timer, and notes linked to readings.
- Implement public review system with spoiler alert and blur.
- Provide annual reading goal definition and tracking with progress visualization and history dashboard.
- Provide public profiles with optional fields (favorite books, contact methods) and simple reader discovery, without aggregated feed or active interactions in the MVP.

## 5. Functional Requirements

### Authentication
- **RF01** — Register new users with email and password.
- **RF02** — Authenticate users via email and password.
- **RF03** — Store passwords using bcrypt.
- **RF04** — Maintain authenticated session via JWT.
- **RF05** — Logout, invalidating client-side session token.
- **RF06** — Password recovery via email token.

### Catalog
- **RF07** — Book search by title, author, or ISBN.
- **RF08** — Integrate Google Books API to retrieve book data (title, author, cover, synopsis, ISBN, genres, total pages).
- **RF09** — Persist book data locally for already-consulted titles, prioritizing local lookup before external API.
- **RF10** — Manual book registration when external search fails.

### Library (was "Estante" in Portuguese, renamed to "Biblioteca" / Library)
- **RF11** — Classify a book as "Want to Read", "Reading", "Read", or "Abandoned".
- **RF12** — Change book status at any time.
- **RF13** — Display the user's library organized by status, with filters and ordering by title, author, date added.
- **RF14** — Automatically record completion date when status changes to "Read".

### Reading Journal
- **RF15** — Register reading sessions via active timer (start, end, capture initial and final page).
- **RF16** — Register reading sessions after the fact (manual input of initial page, final page, date).
- **RF17** — Calculate and update current page based on the most recent session's final page.
- **RF18** — Register private notes linked to a reading, with optional reference field (free text: page, range, or chapter).
- **RF19** — Edit and delete sessions and notes by their owner.
- **RF20** — Notes visible only to their owner.

### Reviews
- **RF21** — Allow user to review a book with status "Read" via numeric rating and text review.
- **RF22** — Display reviews publicly on the book page.
- **RF23** — Allow user to mark review as containing spoilers.
- **RF24** — Hide spoiler-marked reviews with blur, revealing content only after explicit "Show review" button click.
- **RF25** — Allow author to edit and delete own reviews.

### Goals
- **RF26** — Allow user to set numeric annual reading goal, limited to one goal per year.
- **RF27** — Display progress bar showing user advancement vs. current goal.
- **RF28** — Display reading history charts (most-read genres, pages read over time).

### Profile
- **RF29** — Allow user to personalize profile with optional fields (favorite books, contact methods).

### Privacy
- **RF30** — Allow user to request permanent account deletion, removing personal data and preserving public reviews with anonymized authorship, in compliance with LGPD (Law 13.709/2018).
- **RF31** — Record explicit user consent to terms of use and privacy policy at registration.

### Lists
- **RF32** — Apply pagination to potentially long lists (library, sessions, reviews).

### Presentation
- **RF33** — The home screen, after authentication, must display books in "Reading" status with shortcut for new session registration and visual indication of current annual goal progress.

## 6. Non-Functional Requirements

### Performance
- **RNF01** — API must log response time of each request in structured logs for later performance analysis.

### Security
- **RNF02** — Passwords stored using bcrypt.
- **RNF03** — Authentication based on JWT with short expiration time, accompanied by refresh token mechanism.
- **RNF04** — All private API routes must require valid token.
- **RNF05** — Client-server communication exclusively via HTTPS in production.
- **RNF06** — API must apply rate limiting per IP and per authenticated user, returning HTTP 429 when exceeded.

### Maintainability
- **RNF07** — Backend must follow separation of concerns in layers (handlers, services, repositories).
- **RNF08** — Sensitive configurations (DB credentials, API keys, JWT secrets) must be handled via environment variables.
- **RNF09** — Repository versioned via Git, history organized per functionality, following Conventional Commits.
- **RNF10** — Database schema evolution controlled via versioned migrations tool.

### Observability
- **RNF11** — Application must produce structured logs (JSON format) with levels (info, warn, error) via Go's standard `log/slog` package.

### Operations
- **RNF12** — Application must expose `/health` endpoint for orchestrator availability check.
- **RNF13** — Application must implement graceful shutdown, terminating in-flight connections before process exit.

### Database
- **RNF14** — PostgreSQL must use referential integrity constraints to ensure consistency.
- **RNF15** — Critical operations involving multiple tables (reading session registration, account deletion) must execute within transactions.

### Integrations
- **RNF16** — Failures in Google Books API must be handled with appropriate user message, without compromising rest of application's availability.
- **RNF17** — Google Books API requests must respect platform quota limits, mitigated by local cache.
- **RNF18** — Book search layer must be implemented via interface, allowing replacement or addition of external sources in the future.

### Usability
- **RNF19** — Interface must be responsive, adapting to mobile and desktop devices.
- **RNF20** — Error messages to users must be clear and objective, without exposing technical internal details.

### Privacy
- **RNF21** — Personal data must be handled per Law 13.709/2018 (LGPD), including right to deletion, principle of minimal collection, and explicit consent recording.

## 7. Conceptual Data Model

### Entities and Business Attributes

| Conceptual entity (Portuguese) | Code/DB name (English) | Business attributes |
|---|---|---|
| Usuário | User / `users` | name, email, password (hash), registration date, favorite books (optional), contact methods (optional), consent timestamp |
| Livro | Book / `books` | title, author, genres (array), ISBN, synopsis, cover, total pages, source (Google Books / manual) |
| Leitura | Reading / `readings` | status (Want to Read / Reading / Read / Abandoned), current page, date added, completion date |
| Sessão | Session / `sessions` | initial page, final page, session date, duration (when registered with timer) |
| Anotação | Note / `notes` | content, optional reference (free text: page, range, or chapter), creation date |
| Resenha | Review / `reviews` | rating, text, spoiler flag, creation date |
| Meta | Goal / `goals` | year, target book count |

### Main Relationships

- A **User** maintains several **Readings** (1:N); each Reading belongs to exactly one User. The link between User and Book is always mediated by the Reading entity — there is no direct relationship between User and Book.
- A **Book** can appear in several **Readings** (across different users); each Reading references exactly one Book.
- A **Reading** can register several **Sessions** over time (1:N), capturing the reading history.
- A **Reading** can contain several private **Notes** (1:N); each Note may have an optional reference in free text indicating the page, range, or chapter.
- A **User** can write several **Reviews**; a **Book** can receive Reviews from multiple Users, with the constraint that each User can review each Book at most once.
- A **User** can define several **Goals** over time, with the constraint of exactly one Goal per year.
- A **User** can optionally **follow** other Users (reflexive relationship), forming a personal list of followed readers, with no active notification or interaction mechanisms. This is low-priority for the MVP.

### Genre Modeling Decision

Genres are stored as a Postgres `TEXT[]` (native array) directly in the `books` table, **not** as a separate normalized table. Rationale: Google Books returns genres as inconsistent strings, and a separate `genres` + `book_genres` join table adds complexity without clear benefit at this scale. Filtering by genre is preserved via Postgres array operators. If consistency becomes a real pain point, promoting to a separate table is a straightforward refactor. See `docs/DECISIONS/0002-postgres-arrays-for-genres.md`.

### "Current page" Derivation

The `current_page` attribute on `Reading` is conceptually present in the model but **should be derived at query time** from `MAX(final_page)` of associated sessions, not persisted as a column. Persisting it creates a denormalization that requires synchronized updates whenever a session is created, edited, or deleted — a common source of bugs.

### Review Authorization

Although the conceptual model does not show a direct relationship between `Review` and `Reading`, the rule "user can only review a book with status 'Read'" (RF21) is enforced at the application service layer: before creating a `Review`, the service checks for existence of a `Reading` row with `status='read'` for that `(user_id, book_id)` pair. This is a service-level constraint, not expressible as a database foreign key.

## 8. Architecture (high-level)

Three-tier web application orchestrated locally via Docker Compose:

- **SPA Frontend**: React + Vite + React Router + Tailwind CSS + Recharts + Axios, in TypeScript. Communicates with backend via HTTP REST + JSON. In development, served by Vite dev server on port 3000. In production, built as static files served by a CDN-backed host.
- **API Backend**: Go (1.26+) using stdlib `net/http` augmented with the Chi router for routing and middleware. Layered as handlers / services / repositories (RNF07). JWT auth (RNF03), bcrypt hashing (RNF02). Exposes `/health` (RNF12), implements graceful shutdown (RNF13), produces JSON logs via `log/slog` (RNF11). Listens on port 8080.
- **Database**: PostgreSQL 16 (alpine image in dev), accessed only by the backend, with referential constraints (RNF14) and transactional critical operations (RNF15). Schema versioned via golang-migrate (RNF10). Listens on port 5432 internally; in dev exposed on 5433 to avoid conflict with locally-installed Postgres on the developer's machine.

External integration:
- **Google Books API**: called only by the backend via HTTPS for book metadata lookup. Not called from the frontend (avoids quota leakage and centralizes error handling — RNF16, RNF17). Implementation behind a `BookSearcher` interface to allow alternate sources in the future (RNF18).

## 9. Deployment Strategy (preliminary)

Three components hosted on free-tier services (final providers to be confirmed during deploy phase):

- **Frontend** (static SPA): Vercel, Netlify, or Cloudflare Pages. Preference: **Vercel** for Vite simplicity.
- **Backend** (Go containerized): Fly.io, Render, or Railway. Preference: **Fly.io** for stable free tier and Docker-native deployment.
- **Database** (managed Postgres): Neon or Supabase. Preference: **Neon** for stability and ease.

Stateless deployment: frontend and backend deployed independently, communicate exclusively via HTTP REST. No shared session state.

## 10. Product Decisions Already Made

These are decisions already taken during planning conversations, even where not explicit in the requirements list above:

1. **Module name "Estante" was renamed to "Library" (Portuguese: "Biblioteca")** to avoid the redundancy of "virtual shelf" in a digital context, and to align with the entity name `Reading` (the Library contains your Readings).
2. **Reading Journal entity remains "Diário de Bordo"** in the user-facing UI — strong product term.
3. **Reviews vs. Notes are distinct concepts.** Notes are private, granular, and tied to a Reading. Reviews are public, holistic, and tied to a (User, Book) pair. Different tables, different UX surfaces.
4. **Initial screen displays "Continue Reading"**: cards for books in "Reading" status with a button to record a new session, plus the annual goal progress bar. Not a feed of other users' content.
5. **Following other users is low-priority for MVP** (US16). When implemented, it will be a minimal "list of readers I follow" without notifications, chat, or aggregated feed.
6. **Reviews are public on the book page only**, not aggregated into any feed. Profile pages will list a user's own reviews when implemented.
7. **Booklists (curated collections of books) are out of MVP scope**, registered as future possibility.
8. **AI-generated note synthesis is out of MVP scope**, registered as future possibility.
9. **Account deletion (RF30) anonymizes the review author** rather than deleting the review (Reddit-style "[deleted user]"). Preserves public discourse while honoring deletion request.
10. **Annotation reference is a single free-text field**, not three separate fields for page/range/chapter. The user writes "87", "87-92", "ch. 3", "epilogue" — whatever fits.
11. **First page input on reading sessions**: on the first session of a Reading, the user enters their starting page (because page 1 is rarely the start of chapter 1). Subsequent sessions auto-suggest the previous session's final page as the new starting page, but allow editing.
12. **Both timer and post-session entry coexist from the MVP**: two buttons on the reading card — "Start session with timer" (records duration) and "Log session" (manual page input, no duration). Same `Session` entity, with `duration` nullable.

## 11. Glossary (Portuguese ↔ English)

For every conceptual term in the academic documentation, the corresponding code/database identifier:

| Conceptual (PT, in docs) | Code identifier (EN) | Database table |
|---|---|---|
| Usuário | `User` | `users` |
| Livro | `Book` | `books` |
| Leitura | `Reading` | `readings` |
| Sessão | `Session` | `sessions` |
| Anotação | `Note` | `notes` |
| Resenha | `Review` | `reviews` |
| Meta | `Goal` | `goals` |
| Biblioteca (módulo) | "Library" (module name in code/UI strings) | n/a |
| Diário de Bordo (módulo) | "Reading Journal" (in code), "Diário de Bordo" (in UI) | n/a |
| Quero Ler (status) | `want_to_read` | enum value |
| Lendo (status) | `reading` | enum value |
| Lido (status) | `read` | enum value |
| Abandonado (status) | `abandoned` | enum value |

## 12. Out of Scope (for MVP)

The following are explicitly **not** part of the MVP. They are listed in the academic document under "Possibilidades Futuras" (Future Possibilities):

- Collaborative booklists (book playlists, public discovery, ranking)
- AI-generated synthesis of user notes
- Focus mode with configurable minimum time and reading speed statistics
- Annual "Wrapped" — shareable image summarizing the user's reading year
- Streak system — counter of consecutive reading days
- Followed-users feed of recent reviews

---

**Last updated:** when adding new requirements or changing existing ones, update this file first, then propagate to other docs and code.