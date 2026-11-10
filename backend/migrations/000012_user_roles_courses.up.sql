CREATE TYPE role AS ENUM ('student', 'teacher', 'admin');
CREATE TYPE term AS ENUM ('winter', 'spring', 'summer', 'fall');

ALTER TABLE users
ADD COLUMN role role NOT NULL DEFAULT 'student';

CREATE TABLE IF NOT EXISTS courses (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    start_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '6 months'),
    semester term NOT NULL,
    year INTEGER NOT NULL DEFAULT EXTRACT(YEAR FROM NOW())::INTEGER,
    CONSTRAINT courses_year_range CHECK (year BETWEEN 2024 AND 2124)
);
