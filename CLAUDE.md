# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Running the Application
```bash
go run main.go
```
The server runs on port 8080 with graceful shutdown support.

### Building
```bash
go build -o app .
```

### Docker Build
```bash
docker build -t ohabits .
```

### Database Setup
- PostgreSQL database required
- Schema defined in `schema.sql`
- Uses environment variables: `DATABASE_URL`, `JWT_SECRET`
- Load with `.env` file support via godotenv

## Architecture Overview

**ohabits** is a personal habit tracking web application built with Go, PostgreSQL, and HTMX.

### Core Structure
- **Entry Point**: `main.go` - Sets up database connection and HTTP server with graceful shutdown
- **Server Setup**: `cmd/server/server.go` - Defines all HTTP routes using Gorilla Mux
- **Database Layer**: `internal/db/` - PostgreSQL connection and data access methods
- **API Handlers**: `internal/api/` - Business logic for each feature
- **Frontend Handlers**: `internal/handlers/` - HTTP request handlers that render templates
- **Templates**: `templates/` - HTML templates with HTMX integration
- **Static Assets**: `static/` - CSS, JS, images, and icons

### Key Features
- **User Authentication**: JWT-based with bcrypt password hashing
- **Habit Tracking**: Create habits with scheduled days, track daily completions
- **Workout Management**: Create workout plans with exercises, log workout sessions
- **Daily Logging**: Notes, mood ratings (1-10 scale), and todo items
- **Profile Management**: User profiles with avatar upload support

### Database Schema
- `users` - User accounts with email/password auth
- `habits` - Habit definitions with JSONB scheduled_days
- `habits_completions` - Daily habit completion tracking
- `workouts` - Workout plans with JSONB exercises
- `workout_logs` - Workout session logs with JSONB data
- `todos` - Daily todo items
- `notes` - Daily text notes
- `mood_ratings` - Daily mood ratings (1-10)

### Authentication Flow
- Protected routes use `util.AuthMiddleware`
- JWT tokens for session management
- Login/signup pages for unauthenticated users
- User context passed through middleware

### Frontend Architecture
- Server-side rendered HTML templates
- HTMX for dynamic interactions without full page reloads
- Responsive design with custom CSS
- Partials for reusable components (calendar, habit items, etc.)

### Key Dependencies
- **gorilla/mux**: HTTP routing
- **pgx/v5**: PostgreSQL driver and connection pooling
- **golang-jwt/jwt**: JWT token handling
- **golang.org/x/crypto**: Password hashing
- **google/uuid**: UUID generation
- **golang.org/x/image**: Image processing for avatars