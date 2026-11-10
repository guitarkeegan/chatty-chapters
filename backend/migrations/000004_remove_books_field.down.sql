-- Down migration

-- Add the books column back to the users table
ALTER TABLE users
ADD COLUMN books TEXT[] NOT NULL;

-- Drop the user_books table
DROP TABLE IF EXISTS user_books;
