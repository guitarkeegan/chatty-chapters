# Chatty Chapters API

The Go API behind Chatty Chapters. See the [root README](../README.md) for what the
app does and how the pieces fit together.

## Architecture

| Area | Details |
| --- | --- |
| Language | Go, `net/http` with `julienschmidt/httprouter` |
| Data | Postgres 16 with `pgvector` and `citext`, accessed with `pgx` and versioned SQL migrations (`golang-migrate`) |
| AI | OpenAI `gpt-4o` and `gpt-4o-mini` through `langchaingo`; `text-embedding-3-large` embeddings; tool calling for structured grading |
| Auth | AWS Cognito; the API verifies ID tokens against Cognito's JWKS (cached), stores tokens in cookies, and refreshes them transparently in middleware; `student` and `teacher` roles |
| Realtime | `gorilla/websocket` for streaming chat |
| Deployment (when it was live) | A single EC2 instance with systemd and Caddy for TLS; see `remote/` |

## Project layout

```
cmd/api/          HTTP server: routes, middleware, handlers, auth
internal/ai/      LLM chat, quiz generation, and grading
internal/data/    Postgres models
internal/seed/    demo data and the book embedding pipeline
internal/canvas/  Canvas LMS API client for syncing grades (work in progress, not yet wired in)
migrations/       SQL migrations
books/            chapter text, one file per chapter
dummy_client/     a minimal HTML client for testing login without the frontend
remote/           server setup script, systemd unit, and Caddyfile
```

## Running locally

You'll need Go 1.22+, Docker, and an OpenAI API key. Logging in also requires a
Cognito user pool.

```sh
cp .env.example .env   # then add your OPENAI key (and Cognito ids, to log in)
make up                # start Postgres + pgvector in Docker and run the migrations
make seed              # demo users, books, courses, and the book embeddings
make run/front         # start the API on :8080 with CORS open to the local frontend
```

Run `make help` to see every target.

### Logging in

Auth goes through Cognito, so you'll need your own user pool and app client (set
`USER_POOL_ID` and `USER_POOL_CLIENT_ID`). The seeded demo users have random
`cognito_sub` values; to log in as one, set its `cognito_sub` to the `sub` of a user
in your pool. Make the teacher a teacher with
`UPDATE users SET role = 'teacher' WHERE email = 'teacher@example.com';`.

## Tests

```sh
make test               # unit tests, no services needed (this is what CI runs)
make test/integration   # also runs tests that need Postgres, OpenAI, Cognito, and Canvas
```

## Sample books

The real courses used copyrighted textbooks, which aren't included. The demo uses two
public domain books from [Project Gutenberg](https://www.gutenberg.org):

- *Alice's Adventures in Wonderland* by Lewis Carroll
- *Life of Chopin* by Franz Liszt, translated by M. Walker Cook

Each chapter is a text file in `books/<title>/`, named
`title-author-edition-chapter.txt`. The embedding seeder parses that name into the
metadata used to filter search results by book and chapter. To add a book, add its
files there and its chapters in `internal/seed/seed_chapters.go`.

## Frontend

The Next.js frontend is in [`../frontend`](../frontend).
