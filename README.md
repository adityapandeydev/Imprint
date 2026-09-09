# Imprint

> A modern book discovery, cataloging, and personal reading collection application.

Imprint is designed for readers who love books. It helps you discover books naturally by title, author, or ISBN, inspect complete edition metadata, and organize your personal reading life (`Want to Read`, `Currently Reading`, `Finished`, `Abandoned`).

Architecturally, Imprint is engineered as the foundation for an eventual edition-level comparison platform ("Trivago for books")—decoupling the intellectual **Work** from the physical or digital **Edition** (ISBN/ASIN) so multi-retailer price and availability comparison can be introduced seamlessly.

---

## Tech Stack

* **Backend:** Go 1.22+ (Standard library HTTP routing / Chi, `pgxpool`, structured `slog`)
* **Database:** PostgreSQL 16+ (Optimized for [Neon Serverless](https://neon.tech))
* **Frontend:** React 19, TypeScript, Vite, Tailwind CSS
* **Package Manager & Task Runner:** [Bun](https://bun.sh) / `make`
* **External Providers:** Open Library API, Google Books API

---

## Getting Started

### Prerequisites

* [Go 1.22+](https://go.dev/dl/)
* [Bun](https://bun.sh)
* A [Neon PostgreSQL](https://neon.tech) database (or any local PostgreSQL instance)

### 1. Clone & Configure Environment

```bash
cp .env.example .env
```

Open `.env` and paste your Neon PostgreSQL connection string:
```env
DATABASE_URL=postgres://user:password@ep-project.aws.neon.tech/neondb?sslmode=require
```

### 2. Run Using Bun or Make

Imprint provides root tasks configured for Bun (convenient on Windows):

```bash
# Run the Go backend server
bun run dev:backend

# Run the React frontend dev server
bun run dev:frontend

# Run backend tests
bun run test:backend

# Run frontend tests
bun run test:frontend
```

*(Or use `make dev-backend`, `make dev-frontend`, `make test-backend` if you have `make` installed).*

---

## Project Structure

```text
imprint/
├── backend/            # Go backend (API transport, application services, domain models)
├── frontend/           # React 19 + TypeScript + Vite frontend
├── .env.example        # Configuration template
├── Makefile            # Standard POSIX task runner
├── package.json        # Bun task scripts
├── LICENSE             # MIT License
└── README.md
```

---

## License

MIT
