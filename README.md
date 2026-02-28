# hello-devops-world

A Go-based API sandbox for practicing advanced DevOps engineering and Infrastructure as Code.

---

## Project Overview

This repository serves as a foundational project for my **Zero-to-Expert DevOps journey**.
It contains a lightweight, production-ready Golang REST API connected to a PostgreSQL database, containerized using industry best practices.

The goal of this sandbox is to progressively implement and practice:

* CI/CD pipelines
* Kubernetes orchestration
* Terraform provisioning
* Observability & monitoring

---

## Tech Stack & DevOps Practices

### Backend

* **Golang** – RESTful API

### Database

* **PostgreSQL 15**

### Observability

* **Prometheus metrics** exposed at `/metrics`

### Containerization (Docker)

* Multi-stage builds (minimal image size)
* Statically linked binary
* Non-root execution (`appuser`)
* Proper PID 1 handling (`dumb-init`)
* Graceful shutdown support

### Orchestration (Docker Compose)

* Public/Private bridge networks
* Service health checks (`service_healthy`)
* Startup order dependency management
* Persistent volume for PostgreSQL

---

## Architecture

```
Client
   ↓
Go API (Port 8080)
   ↓
PostgreSQL 15
   ↓
Prometheus metrics (/metrics)
```

---

## Quick Start

### Clone the repository

```bash
git clone https://github.com/DiaBmond/hello-devops-world.git
cd hello-devops-world
```

### Spin up the infrastructure

```bash
docker compose up --build -d
```

> Requires Docker v2 (recommended).
> If using legacy CLI: `docker-compose up --build -d`

### Verify services are running

```bash
docker ps
```

You should see containers with status:

```
Up (healthy)
```

---

## Testing the Endpoints

Once the containers are healthy, test the APIs:

---

### Application Health (Used for Probes)

```bash
curl -i http://localhost:8080/health
```

---

### Create a New User

```bash
curl -i -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "user1",
    "email": "user1@example.com"
  }'
```

---

### Fetch All Users

```bash
curl -i http://localhost:8080/users
```

---

### Fetch User by ID

```bash
curl -i http://localhost:8080/users/1
```

---

### View Prometheus Metrics

```bash
curl http://localhost:8080/metrics
```

---

## Environment Variables

| Variable    | Description       |
| ----------- | ----------------- |
| DB_HOST     | PostgreSQL host   |
| DB_PORT     | PostgreSQL port   |
| DB_USER     | Database username |
| DB_PASSWORD | Database password |
| DB_NAME     | Database name     |

---
