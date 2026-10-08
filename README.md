# URL Shortener

A lightweight URL shortener service built with Go, MongoDB, Redis, and Kafka. It supports user registration and login, short URL generation, authenticated URL listing, and redirect tracking with click events.

## Features

- User registration and login with password hashing
- JWT based authentication with access and refresh tokens
- URL shortening and listing for authenticated users
- Redirect handling with Redis caching for faster lookups
- Click tracking and Kafka-based event publishing for analytics
- MongoDB backed storage and Docker-based local setup

## Tech Stack

- Go 1.27+
- Gin HTTP framework
- MongoDB
- Redis
- Kafka
- Docker and Docker Compose
- JWT authentication

## Project Structure

```text
URL-Shortener/
├── Client/
│   └── .gitkeep
├── Server/
│   ├── controllers/
│   ├── database/
│   ├── kafka/
│   ├── middleware/
│   ├── models/
│   ├── routes/
│   ├── utils/
│   ├── .dockerignore
│   ├── .gitkeep
│   ├── Dockerfile
│   ├── docker-compose.yml
│   ├── go.mod
│   ├── go.sum
│   └── main.go
├── .gitignore
├── README.md
└── LICENSE (if present in your repo)
```

## Architecture Overview

The backend is a Go service that exposes HTTP endpoints via Gin. It:

- stores users and shortened URLs in MongoDB
- caches redirect mappings in Redis
- publishes click events to Kafka after a redirect
- validates requests and protects routes with middleware

## Getting Started

### Prerequisites

- Go 1.27 or newer
- Docker and Docker Compose
- MongoDB, Redis, and Kafka, or use the provided Docker Compose setup

### Clone the repository

```bash
git clone https://github.com/harshit3011/URL-Shortener.git
cd URL-Shortener
```

### Run with Docker Compose

```bash
cd Server
docker compose up --build
```

This starts the application and required services:

- app on port 8081
- MongoDB on port 27017
- Redis on port 6379
- Kafka on port 9092

### Run without Docker

From the Server directory:

```bash
go mod download
go run main.go
```

## Environment Variables

The app uses the following environment variables, which are also configured in `Server/docker-compose.yml`:

```env
PORT=8081
MONGODB_URI=mongodb://localhost:27017/
DATABASE_NAME=URL-Shortener
REDIS_ADDR=localhost:6379
KAFKA_ADDR=localhost:9092
SECRET_ACCESS_KEY=your_access_secret
SECRET_REFRESH_KEY=your_refresh_secret
```

If no `.env` file is present, the server falls back to environment variables and default values.

## API Endpoints

### Authentication

- `POST /register` - register a user
- `POST /login` - login with username or email
- `POST /logout` - logout current user

### URL Management

- `GET /urls` - get all URLs for the authenticated user
- `POST /shorten` - create a shortened URL
- `GET /redirect/:shortcode` - redirect to the original URL

### Root

- `GET /` - basic welcome route

## Example Requests

### Register user

```bash
curl -X POST http://localhost:8081/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "email": "alice@example.com",
    "password": "strongpassword123"
  }'
```

### Login

```bash
curl -X POST http://localhost:8081/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "strongpassword123"
  }'
```

### Shorten a URL

```bash
curl -X POST http://localhost:8081/shorten \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "url": "https://example.com/very/long/path"
  }'
```

### Redirect

```bash
curl -I http://localhost:8081/redirect/000001
```

## Notes

- The repository currently has a `Client` directory placeholder and the backend is the main implementation.
- Redis is used as a short lived cache for redirect lookups.
- Kafka click events can be consumed by downstream analytics or monitoring services.

## Contributing

Contributions are welcome. Feel free to open an issue or create a pull request with improvements, bug fixes, or additional tests.

## License

This project does not currently declare a license in the repository metadata. If you plan to publish or share the code more broadly, consider adding an appropriate open source license.
