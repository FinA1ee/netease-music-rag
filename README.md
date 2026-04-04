# NetEase Music RAG

NetEase Music RAG is a full-stack Retrieval-Augmented Generation (RAG) application that enhances your NetEase Cloud Music daily recommendations. It automatically fetches your daily play recommendations, generates rich descriptions and tags (mood, style, scene) using Google's GenAI, and creates vector embeddings stored in a PostgreSQL database (using `pgvector`). This allows you to perform powerful natural language searches across your customized music library.

## Features

- **Automated Ingestion:** A background cron job easily fetches daily recommendations from NetEase Cloud Music.
- **LLM-Powered Enrichment:** Uses Google's GenAI to understand tracks and append descriptive tags (mood, style, scene) and meaningful textual descriptions.
- **Semantic Search:** Stores descriptors as vector embeddings to enable natural language queries (e.g., "upbeat road trip song that makes me feel happy").
- **Modern Tech Stack:** Go (chi router) back-end, React (TypeScript) front-end, and PostgreSQL + pgvector for vector storage.

## Architecture & Tech Stack

- **Frontend:** React + TypeScript UI built with Webpack.
- **Backend:** Go REST API using `go-chi`. Integrates with Google GenAI for LLM completions/embeddings.
- **Database:** PostgreSQL with the `pgvector` extension.
- **Third-party Services:** Uses the `binaryify/netease_cloud_music_api` docker image to bridge requests to NetEase.

## Prerequisites

- [Docker](https://www.docker.com/) and Docker Compose
- [Go 1.25+](https://go.dev/)
- [Node.js](https://nodejs.org/) & npm/yarn (for local frontend development)
- Google Gemini API Key (for `genai`)

## Getting Started

### 1. Launch Services

Launch the supporting backend services (Postgres with `pgvector` and the NetEase Cloud Music API) via Docker Compose:

```bash
docker-compose up -d
```

*(This will provision the database via `./backend/sql/init.sql` and expose the NetEase API on port 3000)*

### 2. Backend Setup

Configure your environment variables. Ensure you have your `GEMINI_API_KEY` (or equivalent) configured for GenAI text generation and embeddings. You may want to create a `.env` file if required by your loaded environment variables.

To run the Go backend server:

```bash
cd backend
go mod tidy
go run ./cmd/server
```

### 3. Frontend Setup

Move into the `frontend` directory, install dependencies, and start the development server.

```bash
cd frontend
npm install
npm run start
```
To build for production, run `npm run build`.

## API Documentation

The backend exposes endpoints to manually trigger ingestion and search your vector database using natural language. 

For full details, see the [API Documentation](./README-API.md).

- `GET /api/search` - Search songs via natural language.
- `POST /api/trigger-job` - Manually trigger the daily fetch & vectorize job.
