# Road map

## Back-end

**API**
    - [ ] Habits :-
        - [ ] GET /api/habits
        - [ ] POST /api/habits/{id}
        - [ ] DELETE /api/habits/{id}

        - [ ] GET /api/habits/{date}
        - [ ] POST /api/habits/{date}

    - [ ] Workout :-
        - [ ] GET /api/workout
        - [ ] GET /api/workout/{id}
        - [ ] POST /api/workout/{id}
        - [ ] DELETE /api/workout/{id}

        - GET /api/workout/{date}
        - POST /api/workout/{date}

    - [ ] TODO :-
        - [ ] GET /api/todo/{date}
        - [ ] POST /api/todo/{date}

    - Note :-
        - [ ] GET /api/note/{date}
        - [ ] POST /api/note/{date}
        - [ ] DELETE /api/note/{date}

    - Rate :-
        - [ ] GET /api/rate/{date}
        - [ ] POST /api/rate/{date}

    - View mode :-
        - [ ] GET /api/view/{month}

    - User :-
        - GET /api/profile/{id}
        - POST /api/profile/{id}

**SQL**
    - [ ] User
    - [ ] habits
        - id
        - user_id
        - name
        - scheduled_days
        - created_at
        - updated_at
    - [ ] habits_completions
        - id
        - habit_id
        - user_id
        - completed
        - date
        - created_at
        - updated_at
    - [ ] workouts
        - id
        - user_id
        - name
        - day
        - exercises
        - created_at
        - updated_at
    - [ ] workout_log
        - id
        - user_id
        - workout_id
        - completed_exercises
        - cardio
        - weight
        - date
        - note
        - created_at
        - updated_at
    - [ ] todos
        - id
        - user_id
        - text
        - completed
        - date
        - created_at
        - updated_at
    - [ ] notes
        - id
        - user_id
        - date
        - content
        - created_at
        - updated_at
    - [ ] mood_ratings
        - id
        - user_id
        - rating
        - date
        - created_at
        - updated_at
    - [ ] profiles
        - id
        - display_name
        - avatar_url
        - password
        - created_at
        - updated_at


File Structure :-
```
ohabits_goth/
│── cmd/
│   ├── server/        # Main entry point for the app
│   │   ├── main.go
│── internal/
│   ├── api/           # Handlers for HTTP requests
│   ├── db/            # Database models and queries
│   ├── services/      # Business logic
│   ├── auth/          # Authentication logic (JWT, Sessions, etc.)
│── migrations/        # Database migrations
│── static/            # Static assets (CSS, JS, images)
│── templates/         # HTML templates for HTMX
│── config/            # Configuration files
│── .env               # Environment variables
│── go.mod             # Go module file
│── go.sum             # Go dependencies
│── Makefile           # Automation tasks (build, run, lint, etc.)
│── README.md
```



