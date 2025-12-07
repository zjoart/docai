# AI Document Analysis Service (DocAI)

A Go microservice that ingests documents (PDF, TXT), extracts their text, and utilizes Large Language Models (LLMs) via OpenRouter to generate summaries and extract key metadata.

## 📋 Prerequisites

Ensure you have the following installed on your local machine:

- **Docker & Docker Compose**
- **Go** (for local development)

## ⚙️ Configuration

1. **Clone the repository**:
   ```bash
   git clone <repo-url>
   cd docai
   ```

2. **Environment Setup**:
   Copy the example environment file ([.env.example](.env.example)):
   ```bash
   cp .env.example .env
   ```
   *Note: Update `OPENROUTER_API_KEY` in `.env` if you wish to test actual LLM analysis.*

## 🏃‍♂️ Getting Started

We use a [`Makefile`](Makefile) to orchestrate workflows.

### Start the Application
```bash
make start-app
```

This runs migrations, creates a minio bucket and runs the server
The server will start at `http://localhost:8080`.

### Other Commands
For a list of all available commands, run:
```bash
make help
```

## 📖 API Documentation

The API is documented using OpenAPI 3.0.

- **Swagger UI**: Accessible at `http://localhost:8080/swagger/` when the server is running.
- **Spec File**: Located at [`docs/swagger.yaml`](docs/swagger.yaml).

## 🧪 Testing

The project includes end-to-end integration tests.
```bash
make test
```

### CI
CI runs `make test-ci` on every push using [.github/workflows/ci.yml](.github/workflows/ci.yml). It spins up Postgres and MinIO containers for tests.

> [!TIP]
> **Secrets Configuration**: To fully enable integration tests, add your `.env` variables (e.g., `OPEN_ROUTER_KEY`) to **GitHub Secrets**. The workflow automatically populates the runner's `.env` file from these secrets.
