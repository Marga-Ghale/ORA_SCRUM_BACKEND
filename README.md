# ORA SCRUM Backend

Full-featured Golang REST API backend for the ORA SCRUM project management application.

## Features

- **Authentication**: JWT-based auth with refresh tokens
- **Workspaces**: Multi-workspace support with member management
- **Spaces**: Organize projects into logical groups
- **Projects**: Full project management with team members
- **Sprints**: Agile sprint management with start/complete workflows
- **Tasks**: Complete task lifecycle with status, priority, labels
- **Comments**: Discussion threads on tasks
- **Notifications**: Real-time notification system
- **Cron Jobs**: Scheduled tasks for reminders and maintenance

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin (HTTP Router)
- **Database**: PostgreSQL 16
- **ORM**: Prisma (via prisma-client-go)
- **Cache**: Redis 7
- **Authentication**: JWT with refresh tokens
- **Scheduler**: robfig/cron

## Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/            # HTTP request handlers
│   │   └── middleware/          # Auth middleware
│   ├── config/                  # Configuration management
│   ├── cron/                    # Scheduled jobs
│   ├── models/                  # Request/Response DTOs
│   ├── notification/            # Notification service
│   ├── repository/              # Database operations
│   └── service/                 # Business logic layer
├── prisma/
│   └── schema.prisma            # Database schema
├── Dockerfile                   # Production container
├── docker-compose.yml           # Container orchestration
└── go.mod                       # Go dependencies
```

## Quick Start

### Using Docker (Recommended)

1. Copy environment file:
```bash
cp .env.example .env
```

2. Start all services:
```bash
docker-compose up -d
```

3. Run database migrations:
```bash
docker-compose --profile migrate up migrate
```

4. (Optional) Open Prisma Studio:
```bash
docker-compose --profile studio up studio
```

The API will be available at `http://localhost:8080`

### Local Development

1. Install Go 1.21+ and Node.js 18+

2. Install dependencies:
```bash
go mod download
npm install -g prisma
```

3. Generate Prisma client:
```bash
go run github.com/steebchen/prisma-client-go generate
```

4. Start PostgreSQL and Redis (via Docker):
```bash
docker-compose up -d db redis
```

5. Run migrations:
```bash
DATABASE_URL="postgresql://postgres:postgres@localhost:5432/ora_scrum?schema=public" prisma migrate deploy
```

6. Start the server:
```bash
go run cmd/api/main.go
```

## API Endpoints

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | Register new user |
| POST | `/api/auth/login` | Login |
| POST | `/api/auth/refresh` | Refresh access token |
| POST | `/api/auth/logout` | Logout |

### Users
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/users/me` | Get current user |
| PUT | `/api/users/me` | Update profile |

### Workspaces
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/workspaces` | List workspaces |
| POST | `/api/workspaces` | Create workspace |
| GET | `/api/workspaces/:id` | Get workspace |
| PUT | `/api/workspaces/:id` | Update workspace |
| DELETE | `/api/workspaces/:id` | Delete workspace |
| GET | `/api/workspaces/:id/members` | List members |
| POST | `/api/workspaces/:id/members` | Add member |
| PUT | `/api/workspaces/:id/members/:userId` | Update role |
| DELETE | `/api/workspaces/:id/members/:userId` | Remove member |
| GET | `/api/workspaces/:id/spaces` | List spaces |
| POST | `/api/workspaces/:id/spaces` | Create space |

### Spaces
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/spaces/:id` | Get space |
| PUT | `/api/spaces/:id` | Update space |
| DELETE | `/api/spaces/:id` | Delete space |
| GET | `/api/spaces/:id/projects` | List projects |
| POST | `/api/spaces/:id/projects` | Create project |

### Projects
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/projects/:id` | Get project |
| PUT | `/api/projects/:id` | Update project |
| DELETE | `/api/projects/:id` | Delete project |
| GET | `/api/projects/:id/members` | List members |
| POST | `/api/projects/:id/members` | Add member |
| DELETE | `/api/projects/:id/members/:userId` | Remove member |
| GET | `/api/projects/:id/sprints` | List sprints |
| POST | `/api/projects/:id/sprints` | Create sprint |
| GET | `/api/projects/:id/tasks` | List tasks |
| POST | `/api/projects/:id/tasks` | Create task |
| GET | `/api/projects/:id/labels` | List labels |
| POST | `/api/projects/:id/labels` | Create label |

### Sprints
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/sprints/:id` | Get sprint |
| PUT | `/api/sprints/:id` | Update sprint |
| DELETE | `/api/sprints/:id` | Delete sprint |
| POST | `/api/sprints/:id/start` | Start sprint |
| POST | `/api/sprints/:id/complete` | Complete sprint |
| GET | `/api/sprints/:id/tasks` | List sprint tasks |

### Tasks
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/tasks/:id` | Get task |
| PUT | `/api/tasks/:id` | Update task |
| PATCH | `/api/tasks/:id` | Partial update |
| DELETE | `/api/tasks/:id` | Delete task |
| PUT | `/api/tasks/bulk` | Bulk update |
| GET | `/api/tasks/:id/comments` | List comments |
| POST | `/api/tasks/:id/comments` | Add comment |

### Comments
| Method | Endpoint | Description |
|--------|----------|-------------|
| PUT | `/api/comments/:id` | Update comment |
| DELETE | `/api/comments/:id` | Delete comment |

### Labels
| Method | Endpoint | Description |
|--------|----------|-------------|
| PUT | `/api/labels/:id` | Update label |
| DELETE | `/api/labels/:id` | Delete label |

### Notifications
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/notifications` | List notifications |
| GET | `/api/notifications/count` | Get counts |
| PUT | `/api/notifications/:id/read` | Mark as read |
| PUT | `/api/notifications/read-all` | Mark all read |
| DELETE | `/api/notifications/:id` | Delete one |
| DELETE | `/api/notifications` | Delete all |

## Cron Jobs

| Schedule | Job | Description |
|----------|-----|-------------|
| Daily 9:00 AM | Due Date Reminders | Notify users of tasks due soon |
| Daily 10:00 AM | Overdue Notifications | Notify users of overdue tasks |
| Daily 9:00 AM | Sprint Ending | Remind of sprints ending soon |
| Weekly Sunday | Cleanup | Remove old read notifications |
| Hourly | Auto-complete | Complete expired sprints |
| Every 30 min | Status Update | Mark inactive users as away |

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `API_PORT` | API server port | 8080 |
| `ENVIRONMENT` | Runtime environment | development |
| `DATABASE_URL` | PostgreSQL connection URL | - |
| `REDIS_URL` | Redis connection URL | redis://localhost:6379 |
| `JWT_SECRET` | JWT signing secret | - |
| `JWT_EXPIRY` | Access token expiry (hours) | 24 |
| `REFRESH_EXPIRY` | Refresh token expiry (days) | 7 |
| `SMTP_*` | Email configuration | - |

## Health Check

```
GET /health
```

Returns `200 OK` with timestamp when healthy.



```

## Step 6: Verify Directory Structure

Your directory should look like:
```
~/Desktop/ORA_SCRUM_BACKEND/
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   └── middleware/
│   ├── config/
│   │   └── config.go       ← Updated
│   ├── cron/
│   ├── db/
│   │   ├── postgres.go     ← NEW
│   │   └── redis.go        ← NEW
│   ├── notification/
│   ├── repository/
│   │   └── repository.go   ← Updated (complete file)
│   ├── seed/
│   │   └── seed.go
│   └── service/
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
└── main.go                  ← Updated

## License

MIT


ORA SCRUM - Deployment Report
Project: ORA SCRUM - Project Management Platform
Date: December 15, 2025
Server: 116.203.47.221 (Ubuntu)
Domain: oratechnologies.io

📋 Executive Summary
Successfully deployed a full-stack application (Go backend + React frontend) with automated CI/CD pipelines, SSL certificates, and Docker containerization on a single Ubuntu server.

🏗️ Architecture Overview
┌─────────────────────────────────────────────────────────────────┐
│                        INTERNET                                  │
└─────────────────────────┬───────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                    CLOUDFLARE DNS                                │
│  scrum.oratechnologies.io ──────► 116.203.47.221                │
│  scrum-api.oratechnologies.io ──► 116.203.47.221                │
└─────────────────────────┬───────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                 SERVER: 116.203.47.221                           │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │                    NGINX (Docker)                          │  │
│  │              Ports: 80, 443 (SSL)                          │  │
│  │  ┌─────────────────────┐  ┌─────────────────────┐         │  │
│  │  │ scrum.oratech...    │  │ scrum-api.oratech...│         │  │
│  │  │    ──► frontend:80  │  │    ──► api:8080     │         │  │
│  │  └─────────────────────┘  └─────────────────────┘         │  │
│  └───────────────────────────────────────────────────────────┘  │
│                          │                  │                    │
│              ┌───────────┴──────┐    ┌──────┴───────┐           │
│              ▼                  ▼    ▼              ▼           │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │    FRONTEND     │  │      API        │  │    DATABASE     │  │
│  │  (React/Vite)   │  │     (Go/Gin)    │  │  (PostgreSQL)   │  │
│  │    Port: 80     │  │   Port: 8080    │  │   Port: 5432    │  │
│  └─────────────────┘  └────────┬────────┘  └─────────────────┘  │
│                                │                                 │
│                                ▼                                 │
│                       ┌─────────────────┐                       │
│                       │     REDIS       │                       │
│                       │   Port: 6379    │                       │
│                       └─────────────────┘                       │
└─────────────────────────────────────────────────────────────────┘

🌐 Live URLs
ServiceURLStatusFrontendhttps://scrum.oratechnologies.io✅ LiveBackend APIhttps://scrum-api.oratechnologies.io✅ LiveHealth Checkhttps://scrum-api.oratechnologies.io/health✅ Active

📁 Server File Structure
/root/ORA_SCRUM/
├── docker-compose.yml          # All service definitions
├── .env                        # Environment variables
├── nginx/
│   ├── nginx.conf              # Main nginx config
│   └── conf.d/
│       └── default.conf        # Site configurations
├── certbot/
│   ├── conf/                   # SSL certificates
│   │   └── live/
│   │       ├── scrum.oratechnologies.io/
│   │       └── scrum-api.oratechnologies.io/
│   └── www/                    # ACME challenge files
└── postgres_data/              # Database volume (persistent)

🐳 Docker Services
docker-compose.yml
yamlservices:
  db:
    image: postgres:16-alpine
    container_name: ora_scrum_db
    restart: unless-stopped
    environment:
      POSTGRES_USER: ${DB_USER:-postgres}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: ${DB_NAME:-ora_scrum}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - ora_network

  redis:
    image: redis:7-alpine
    container_name: ora_scrum_redis
    restart: unless-stopped
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    networks:
      - ora_network

  api:
    image: ghcr.io/${GITHUB_REPO}:latest
    container_name: ora_scrum_api
    restart: unless-stopped
    environment:
      - API_PORT=8080
      - ENVIRONMENT=production
      - DATABASE_URL=postgresql://${DB_USER:-postgres}:${DB_PASSWORD}@db:5432/${DB_NAME:-ora_scrum}?sslmode=disable
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - ora_network

  frontend:
    image: ghcr.io/${GITHUB_FRONTEND_REPO}:latest
    container_name: ora_scrum_frontend
    restart: unless-stopped
    networks:
      - ora_network

  nginx:
    image: nginx:alpine
    container_name: ora_scrum_nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/conf.d:/etc/nginx/conf.d:ro
      - ./certbot/conf:/etc/letsencrypt:ro
      - ./certbot/www:/var/www/certbot:ro
    networks:
      - ora_network

  certbot:
    image: certbot/certbot
    container_name: ora_scrum_certbot
    restart: unless-stopped
    volumes:
      - ./certbot/conf:/etc/letsencrypt
      - ./certbot/www:/var/www/certbot

volumes:
  postgres_data:
  redis_data:

networks:
  ora_network:
    driver: bridge

⚙️ Environment Variables (.env)
env# Server Configuration 
API_PORT=8080
ENVIRONMENT=production

# Database Configuration
DATABASE_URL=postgresql://postgres:postgres@db:5432/ora_scrum?sslmode=disable
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=ora_scrum
DB_PORT=5432

# Redis Configuration
REDIS_URL=redis://redis:6379
REDIS_PORT=6379

# JWT Configuration
JWT_SECRET=<your-secret-key>
JWT_EXPIRY=24
REFRESH_EXPIRY=168

# Email Configuration
SMTP_HOST=server104.webhostnepal.com
SMTP_PORT=456
SMTP_USER=hello@oratechnologies.io
SMTP_PASSWORD=<your-password>
SMTP_FROM=hello@oratechnologies.io
SMTP_FROM_NAME=ORA Scrum
SMTP_USE_TLS=true

# Domain
DOMAIN=oratechnologies.io
EMAIL=admin@oratechnologies.io

# GitHub Repos (lowercase)
GITHUB_REPO=marga-ghale/ora_scrum_backend
GITHUB_FRONTEND_REPO=marga-ghale/ora_scrum_frontend

🔄 CI/CD Pipelines
Backend CI/CD
Repository: Marga-Ghale/ORA_SCRUM_BACKEND
.github/workflows/ci.yml
yamlname: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

env:
  GO_VERSION: '1.23'

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: ora_scrum_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      redis:
        image: redis:7
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Download dependencies
        run: go mod download

      - name: Run tests
        env:
          DATABASE_URL: postgres://postgres:postgres@localhost:5432/ora_scrum_test?sslmode=disable
          REDIS_URL: redis://localhost:6379
          JWT_SECRET: test-secret-key
          ENVIRONMENT: test
        run: go test -v -race ./...

  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [test]
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Build
        run: go build -o ora-scrum-api ./cmd/api
.github/workflows/deploy.yml
yamlname: Deploy

on:
  push:
    branches: [main]

env:
  REGISTRY: ghcr.io

jobs:
  build-and-push:
    name: Build and Push
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set repository name to lowercase
        id: repo
        run: echo "name=$(echo '${{ github.repository }}' | tr '[:upper:]' '[:lower:]')" >> $GITHUB_OUTPUT

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to GitHub Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Build and push Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ env.REGISTRY }}/${{ steps.repo.outputs.name }}:latest

  deploy:
    name: Deploy to Server
    runs-on: ubuntu-latest
    needs: build-and-push

    steps:
      - name: Set repository name to lowercase
        id: repo
        run: echo "name=$(echo '${{ github.repository }}' | tr '[:upper:]' '[:lower:]')" >> $GITHUB_OUTPUT

      - name: Deploy via SSH
        uses: appleboy/ssh-action@v1.0.3
        with:
          host: ${{ secrets.SERVER_HOST }}
          username: ${{ secrets.SERVER_USER }}
          password: ${{ secrets.SERVER_PASSWORD }}
          script: |
            cd /root/ORA_SCRUM
            echo ${{ secrets.GITHUB_TOKEN }} | docker login ghcr.io -u ${{ github.actor }} --password-stdin
            docker pull ghcr.io/${{ steps.repo.outputs.name }}:latest
            docker compose up -d api
            docker image prune -f

      - name: Verify deployment
        run: |
          sleep 15
          response=$(curl -s -o /dev/null -w "%{http_code}" https://scrum-api.oratechnologies.io/health)
          if [ "$response" = "200" ]; then
            echo "✅ Deployment successful!"
          else
            echo "⚠️ API status: $response"
          fi

Frontend CI/CD
Repository: Marga-Ghale/ORA_SCRUM_FRONTEND
.github/workflows/deploy.yml
yamlname: Deploy

on:
  push:
    branches: [main]

env:
  REGISTRY: ghcr.io

jobs:
  build-and-push:
    name: Build and Push
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set repository name to lowercase
        id: repo
        run: echo "name=$(echo '${{ github.repository }}' | tr '[:upper:]' '[:lower:]')" >> $GITHUB_OUTPUT

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to GitHub Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Build and push Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ env.REGISTRY }}/${{ steps.repo.outputs.name }}:latest
          build-args: |
            VITE_API_URL=https://scrum-api.oratechnologies.io

  deploy:
    name: Deploy to Server
    runs-on: ubuntu-latest
    needs: build-and-push

    steps:
      - name: Set repository name to lowercase
        id: repo
        run: echo "name=$(echo '${{ github.repository }}' | tr '[:upper:]' '[:lower:]')" >> $GITHUB_OUTPUT

      - name: Deploy via SSH
        uses: appleboy/ssh-action@v1.0.3
        with:
          host: ${{ secrets.SERVER_HOST }}
          username: ${{ secrets.SERVER_USER }}
          password: ${{ secrets.SERVER_PASSWORD }}
          script: |
            cd /root/ORA_SCRUM
            echo ${{ secrets.GITHUB_TOKEN }} | docker login ghcr.io -u ${{ github.actor }} --password-stdin
            docker pull ghcr.io/${{ steps.repo.outputs.name }}:latest
            docker compose up -d frontend
            docker image prune -f

🔐 GitHub Secrets Required
Add these secrets to BOTH repositories:
Secret NameValueSERVER_HOST116.203.47.221SERVER_USERrootSERVER_PASSWORD<your-server-password>
Note: GITHUB_TOKEN is automatically provided by GitHub Actions.

🐋 Dockerfiles
Backend Dockerfile
dockerfileFROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

FROM alpine:3.19

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata wget

COPY --from=builder /app/main .
COPY --from=builder /app/internal/db/migrations ./internal/db/migrations

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:8080/health || exit 1

CMD ["./main"]
Frontend Dockerfile
dockerfileFROM node:20-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .

ARG VITE_API_URL=https://scrum-api.oratechnologies.io
ENV VITE_API_URL=$VITE_API_URL

RUN npm run build

FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]

🌐 Nginx Configuration
nginx/conf.d/default.conf
nginxserver {
    listen 80;
    server_name scrum-api.oratechnologies.io scrum.oratechnologies.io;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        return 301 https://$host$request_uri;
    }
}

server {
    listen 443 ssl;
    http2 on;
    server_name scrum-api.oratechnologies.io;

    ssl_certificate /etc/letsencrypt/live/scrum-api.oratechnologies.io/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/scrum-api.oratechnologies.io/privkey.pem;

    location / {
        proxy_pass http://api:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

server {
    listen 443 ssl;
    http2 on;
    server_name scrum.oratechnologies.io;

    ssl_certificate /etc/letsencrypt/live/scrum.oratechnologies.io/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/scrum.oratechnologies.io/privkey.pem;

    location / {
        proxy_pass http://frontend:80;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

🔒 SSL Certificates
SSL certificates are managed by Let's Encrypt via Certbot.
DomainCertificate PathExpiryscrum-api.oratechnologies.io/etc/letsencrypt/live/scrum-api.oratechnologies.io/Auto-renewsscrum.oratechnologies.io/etc/letsencrypt/live/scrum.oratechnologies.io/Auto-renews
Manual Renewal (if needed)
bashdocker compose run --rm --entrypoint "" certbot certbot renew
docker compose restart nginx

📊 Cloudflare DNS Configuration
TypeNameContentProxy StatusAscrum116.203.47.221DNS only (grey)Ascrum-api116.203.47.221DNS only (grey)
SSL/TLS Mode: Full (strict)

🛠️ Common Commands
Server Management
bash# SSH to server
ssh root@116.203.47.221

# Navigate to project
cd /root/ORA_SCRUM

# View all containers
docker ps

# View logs
docker compose logs -f api
docker compose logs -f frontend
docker compose logs -f nginx

# Restart services
docker compose restart api
docker compose restart frontend
docker compose restart nginx

# Pull and restart all
docker compose pull
docker compose up -d

# Clean unused images
docker image prune -f
Database
bash# Access PostgreSQL
docker exec -it ora_scrum_db psql -U postgres -d ora_scrum

# Backup database
docker exec ora_scrum_db pg_dump -U postgres ora_scrum > backup.sql

# Restore database
cat backup.sql | docker exec -i ora_scrum_db psql -U postgres -d ora_scrum
Troubleshooting
bash# Check container status
docker ps -a

# View container logs
docker logs ora_scrum_api
docker logs ora_scrum_frontend
docker logs ora_scrum_nginx

# Check nginx config
docker exec ora_scrum_nginx nginx -t

# Restart specific container
docker compose restart <service_name>

📈 Deployment Flow
┌─────────────────────────────────────────────────────────────────┐
│                    DEVELOPER PUSHES CODE                         │
│                      to main branch                              │
└─────────────────────────┬───────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                    GITHUB ACTIONS                                │
│  1. Checkout code                                                │
│  2. Run tests (backend only)                                     │
│  3. Build Docker image                                           │
│  4. Push to ghcr.io                                              │
└─────────────────────────┬───────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                    SSH TO SERVER                                 │
│  1. docker login ghcr.io                                         │
│  2. docker pull <image>:latest                                   │
│  3. docker compose up -d <service>                               │
│  4. docker image prune -f                                        │
└─────────────────────────┬───────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                    HEALTH CHECK                                  │
│  Verify service is responding                                    │
└─────────────────────────────────────────────────────────────────┘

✅ Checklist
Completed

 Server setup with Docker
 PostgreSQL database container
 Redis cache container
 Nginx reverse proxy with SSL
 Let's Encrypt SSL certificates
 Backend CI/CD pipeline
 Frontend CI/CD pipeline
 Cloudflare DNS configuration
 GitHub Container Registry setup
 Automated deployments on push to main

Security Notes

 SSL/TLS encryption enabled
 Cloudflare set to Full (strict) mode
 Database not exposed to public
 Redis not exposed to public
 Environment variables for secrets


📞 Quick Reference
ItemValueServer IP116.203.47.221Frontend URLhttps://scrum.oratechnologies.ioAPI URLhttps://scrum-api.oratechnologies.ioHealth Checkhttps://scrum-api.oratechnologies.io/healthProject Path/root/ORA_SCRUMBackend RepoMarga-Ghale/ORA_SCRUM_BACKENDFrontend RepoMarga-Ghale/ORA_SCRUM_FRONTEND

Report Generated: December 15, 2025
Status: ✅ All Systems Operational


