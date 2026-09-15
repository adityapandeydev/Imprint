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

## High-Scale Search Architecture: MeiliSearch (Rust) Roadmap

As Imprint scales, third-party public APIs (such as Open Library and Google Books) introduce latency and rate-limit bottlenecks. To ensure sub-30ms instantaneous search across 35+ million titles with zero external network reliance, Imprint is engineered with a roadmap to support an embedded, dedicated search engine.

### Why MeiliSearch (Rust)?

* **Rust Performance & Memory Safety:** Built in **Rust**, MeiliSearch delivers C/Rust-level execution speeds with memory safety, zero-cost abstractions, and SIMD-accelerated text tokenization.
* **Instantaneous Search:** Returns relevant, typo-tolerant, as-you-type results in **15ms to 30ms**.
* **Pluggable Architecture:** Imprint's Go backend decouples search via the `domain.BookProvider` interface. A `MeiliSearchProvider` can be plugged in as a drop-in primary provider without altering any API handlers or frontend components.

### Bulk Data Ingestion Pipeline (To-Do / Upcoming)

1. **Open Library Data Dump:** Open Library publishes monthly bulk database dumps (`ol_dump_works.txt.gz`, ~2.2GB compressed, containing ~35M works).
2. **Streaming ETL Ingestion:** A Go batch worker streams and normalizes the TSV dump into clean documents (Title, Authors, Original Year, Subject Tags, Open Library Work ID) and batches them directly into MeiliSearch using its native Go SDK (`github.com/meilisearch/meilisearch-go`).
3. **Hybrid Fallback:** For freshly published books released between monthly dump cycles, the system falls back seamlessly to live Google Books and Open Library API lookups.

### Current Progressive Hybrid Pipeline

To maintain zero server bloat on development machines and stay within serverless database limits, Imprint currently operates a lightweight **Unified Progressive Search Pipeline**:
* **7-Day Persistent Search Cache:** High-frequency and trending searches are cached in Neon PostgreSQL with complete metadata (works, authors, covers, and editions) for **sub-15ms** instant responses.
* **Exact-Match & Cover-Prioritized Relevance Scorer:** Candidates are scored dynamically so the exact matching title with verified cover artwork is guaranteed at **Position #1**.
* **Fast-Path Provider:** Google Books delivers fast initial candidate cards (~350ms), while Open Library enriches missing synopses and high-resolution covers without blocking user interaction.

---

## License

MIT
