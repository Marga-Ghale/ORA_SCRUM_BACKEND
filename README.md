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

## License

MIT
