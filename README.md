# Smart Campus Booking System

A microservice-based backend application that enables AITU students and staff to reserve campus rooms. Built with Go, gRPC, NATS, PostgreSQL, Redis, and Docker.

---

## Architecture

The system consists of **4 microservices** and **1 API Gateway**, all orchestrated with Docker Compose.

```
Client (Browser)
      │
      ▼
 API Gateway  (:8080)
      │  HTTP → gRPC
  ┌───┼───────────────────┐
  ▼   ▼                   ▼
User  Room            Booking
Service Service        Service
  │                       │
  │                    NATS (booking.created
  │                         booking.status_changed)
  │                       │
  │                       ▼
  │               Notification Service
  │                       │
  └───────────────── SMTP Email
```

Each microservice has its own **PostgreSQL database** and follows **Clean Architecture**:

| Layer | Responsibility |
|---|---|
| Domain | Pure Go structs, no dependencies |
| Use Case | Business logic, interface-driven |
| Repository | Database access via GORM |
| Transport | gRPC server, translates proto ↔ domain |

---

## Services

| Service | Port | Responsibility |
|---|---|---|
| API Gateway | 8080 | HTTP entry point, JWT auth, HTML frontend |
| User Service | 50051 | Registration, login, JWT issuance |
| Room Service | 50052 | Room CRUD, Redis caching |
| Booking Service | 50053 | Booking lifecycle, conflict detection |
| Notification Service | 50054 | NATS consumer, email sending |

---

## Tech Stack

- **Language:** Go 1.23+
- **RPC:** gRPC + Protocol Buffers
- **Web Framework:** Gin
- **ORM:** GORM
- **Database:** PostgreSQL (Supabase)
- **Cache:** Redis 7
- **Message Queue:** NATS
- **Email:** net/smtp + MailerSend
- **Observability:** Prometheus + Grafana
- **Containerization:** Docker + Docker Compose

---

## Getting Started

### Prerequisites

- Docker and Docker Compose installed
- Git

### Clone and Run

```bash
git clone https://github.com/Aruzhan38/SmartCampusBookingSystem.git
cd SmartCampusBookingSystem
docker compose up --build
```

The API Gateway will be available at `http://localhost:8080`.

### Environment Variables

Each service reads its config from environment variables defined in `docker-compose.yml`. Key variables:

| Variable | Description |
|---|---|
| `DATABASE_URL` | PostgreSQL connection string |
| `REDIS_ADDR` | Redis address (e.g. `redis:6379`) |
| `NATS_URL` | NATS connection URL |
| `JWT_SECRET` | Secret key for JWT signing |
| `SMTP_HOST` | SMTP server host |
| `SMTP_USER` | SMTP username |
| `SMTP_PASS` | SMTP password |
| `USER_SERVICE_ADDR` | gRPC address of User Service |

---

## API Reference

Base URL: `http://localhost:8080`

### Authentication

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| POST | `/api/register` | ❌ | Register a new user |
| POST | `/api/login` | ❌ | Login, returns JWT token |

**Register body:**
```json
{
  "full_name": "Aruzhan Toktarbekova",
  "email": "aruzhan@aitu.edu.kz",
  "password": "123456",
  "role": "Student"
}
```
> `role`: `Student` / `Professor` / `Admin`

**Login body:**
```json
{
  "email": "aruzhan@aitu.edu.kz",
  "password": "123456"
}
```
Returns: `{ "token": "eyJ..." }`

For all protected endpoints, include the header:
```
Authorization: Bearer <token>
```

---

### Rooms

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| GET | `/rooms` | ❌ | List all rooms (Redis cached) |
| GET | `/rooms/search?capacity=30` | ❌ | Search rooms by min capacity |
| GET | `/rooms/:id` | ✅ | Get room by ID |
| POST | `/rooms` | ✅ Admin | Create a room |
| PUT | `/rooms/:id` | ✅ Admin | Update a room |

**Create / Update room body:**
```json
{
  "room_number": "B401",
  "capacity": 45,
  "building_id": "1",
  "description": "Lecture room"
}
```

---

### Bookings

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| POST | `/bookings` | ✅ | Create a booking |
| GET | `/bookings/my` | ✅ | List own bookings (Admin: all bookings) |
| GET | `/bookings/:id` | ✅ | Get booking by ID |
| DELETE | `/bookings/:id` | ✅ | Cancel a booking |
| PATCH | `/bookings/:id/status` | ✅ Admin | Update booking status |

**Create booking body:**
```json
{
  "room_id": 1,
  "start_time": "2026-05-25T10:00:00Z",
  "end_time": "2026-05-25T11:00:00Z",
  "purpose": "Study session"
}
```
> `user_id` is extracted automatically from the JWT token.
> Times must be in **RFC3339** format.

**Update status body:**
```json
{
  "status": "CONFIRMED"
}
```
> `status`: `CONFIRMED` / `CANCELLED` / `REJECTED`

---

### Users

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| GET | `/users/:id` | ✅ | Get user by ID |

---

### Monitoring

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| GET | `/metrics` | ❌ | Prometheus metrics |

Grafana is available at `http://localhost:3000` (default credentials: `admin` / `admin`).

---

## gRPC Endpoints

The system exposes **17 gRPC endpoints** across 4 services. All definitions are in the shared proto repository.

| # | Service | Method | Description |
|---|---|---|---|
| 1 | User | RegisterUser | Register a new user |
| 2 | User | LoginUser | Authenticate and return JWT |
| 3 | User | GetUserById | Get user profile by ID |
| 4 | User | ValidateToken | Validate JWT, return user ID and role |
| 5 | Room | CreateRoom | Add a new room |
| 6 | Room | GetRoomById | Get room by ID |
| 7 | Room | ListRooms | List all rooms (Redis cached) |
| 8 | Room | UpdateRoom | Update room, invalidate cache |
| 9 | Room | SearchAvailableRooms | Find rooms by min capacity |
| 10 | Booking | CreateBooking | Create booking with conflict check |
| 11 | Booking | GetBookingById | Get booking by ID |
| 12 | Booking | ListUserBookings | List bookings for a user |
| 13 | Booking | CancelBooking | Cancel a booking |
| 14 | Booking | UpdateBookingStatus | Update booking status (admin) |
| 15 | Notification | SendNotification | Send and persist a notification |
| 16 | Notification | GetNotificationsByUser | Get all notifications for a user |
| 17 | Notification | MarkNotificationAsRead | Mark notification as read |

---

## Event Flow (NATS)

The system publishes two events:

**`booking.created`** — published when a booking is successfully created. Triggers a confirmation email to the user.

**`booking.status_changed`** — published when an admin updates booking status. Triggers a status update email (`approved` or `rejected`).

```
Booking Service  →  NATS  →  Notification Service  →  Email (SMTP)
```

---

## Frontend

The API Gateway serves a built-in web frontend via HTML templates. No separate server needed.

| Page | URL |
|---|---|
| Dashboard | `/` |
| Login | `/login` |
| Register | `/register` |
| Rooms | `/rooms-ui` |
| My Bookings | `/bookings-ui` |
| Profile | `/profile-ui` |
| Admin Panel | `/admin-ui` |

All pages are protected by JWT middleware. Admin-only pages redirect non-admin users automatically.

---

## Running Tests

```bash
# Run all tests across all services
cd booking-service && go test ./...
cd notification-service && go test ./...
cd room-service && go test ./...
cd user-service && go test ./...
```

Tests use in-memory fake repositories — no database or infrastructure required.

---

## Project Structure

```
SmartCampusBookingSystem/
├── api-gateway/
│   ├── cmd/apiGateway/main.go
│   └── internal/
│       ├── client/          # gRPC clients
│       ├── config/
│       ├── metrics/         # Prometheus middleware
│       ├── middleware/       # JWT auth
│       └── transport/http/  # Handlers + HTML templates
├── booking-service/
│   └── internal/
│       ├── domain/
│       ├── messaging/       # NATS publisher
│       ├── repository/
│       ├── transport/grpc/
│       └── usecase/
├── notification-service/
│   └── internal/
│       ├── mail/            # SMTP sender
│       ├── messaging/       # NATS consumer
│       ├── repository/
│       └── usecase/
├── room-service/
│   └── internal/
│       ├── cache/           # Redis cache
│       ├── repository/
│       └── usecase/
├── user-service/
│   └── internal/
│       ├── repository/
│       └── usecase/
├── monitoring/
│   └── prometheus.yml
└── docker-compose.yml
```

---