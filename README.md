# AI Document Analysis Service (DocAI)

A high-performance Go microservice that ingests documents (PDF, TXT), extracts their text, and utilizes Large Language Models (LLMs) via OpenRouter to generate summaries and extract key metadata.

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
   Copy the example environment file:
   ```bash
   cp .env.example .env
   ```
   *Note: Update `OPENROUTER_API_KEY` in `.env` if you wish to test actual LLM analysis.*

## 🏃‍♂️ Getting Started

We use a `Makefile` to orchestrate workflows.

### Start the Application (Full Stack)
```bash
make start-app
```
The server will start at `http://localhost:8080`.

### Other Commands
For a list of all available commands, run:
```bash
make help
```

## 📖 API Documentation

The API is documented using OpenAPI 3.0.

- **Swagger UI**: Accessible at `http://localhost:8080/swagger/` when the server is running.
- **Spec File**: Located at `docs/swagger.yaml`.

## 🧪 Testing

The project includes end-to-end integration tests.
```bash
make test
```
