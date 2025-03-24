-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- User entity
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    display_name TEXT NOT NULL,
    avatar_url TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Habit entity
CREATE TABLE habits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    scheduled_days JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE habits_completions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    habit_id UUID REFERENCES habits(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    completed BOOLEAN NOT NULL,
    date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Workout entity
CREATE TABLE workouts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    day TEXT NOT NULL,
    exercises JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE workout_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    completed_exercises JSONB NOT NULL,
    cardio JSONB NOT NULL,
    weight FLOAT NOT NULL,
    date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Todo entity
CREATE TABLE todos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    completed BOOLEAN NOT NULL,
    date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Notes entity
CREATE TABLE notes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    text TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Mood ratings entity
CREATE TABLE mood_ratings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    rating INT NOT NULL CHECK (rating BETWEEN 1 AND 10),
    date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Insert dummy data
INSERT INTO users (password, display_name, avatar_url) VALUES
    ('12345', 'most3mr', 'https://example.com/avatar1.png'),
    ('12345', 'test', 'https://example.com/avatar2.png');

INSERT INTO habits (user_id, name, scheduled_days) VALUES
    ((SELECT id FROM users LIMIT 1), 'Morning Run', '["Monday", "Wednesday", "Friday"]'::jsonb),
    ((SELECT id FROM users LIMIT 1 OFFSET 1), 'Read a book', '["Tuesday", "Thursday"]'::jsonb);

INSERT INTO habits_completions (habit_id, user_id, completed, date) VALUES
    ((SELECT id FROM habits LIMIT 1), (SELECT id FROM users LIMIT 1), true, '2025-03-22');

INSERT INTO workouts (user_id, name, day, exercises) VALUES
    ((SELECT id FROM users LIMIT 1), 'Push Day', 'Monday', '[{"name": "Bench Press", "sets": 4, "reps": 8}]'::jsonb);

INSERT INTO workout_logs (user_id, workout_id, completed_exercises, cardio, weight, date, note) VALUES
    ((SELECT id FROM users LIMIT 1), (SELECT id FROM workouts LIMIT 1), '[{"name": "Bench Press", "sets": 4, "reps": 8}]'::jsonb, '[{"type": "Treadmill", "duration": 15}]'::jsonb, 80.5, '2025-03-20', 'Felt great today!');

INSERT INTO todos (user_id, text, completed, date) VALUES
    ((SELECT id FROM users LIMIT 1), 'Buy groceries', false, '2025-03-21');

INSERT INTO notes (user_id, date, text) VALUES
    ((SELECT id FROM users LIMIT 1), '2025-03-21', 'Meeting notes from today.');

INSERT INTO mood_ratings (user_id, rating, date) VALUES
    ((SELECT id FROM users LIMIT 1), 4, '2025-03-22');
