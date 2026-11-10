-- Drop the courses table first because it depends on the semester type
DROP TABLE IF EXISTS courses;

-- Remove the role column from the users table
ALTER TABLE users DROP COLUMN role;

-- Drop the enum types after removing dependencies
DROP TYPE IF EXISTS role;
DROP TYPE IF EXISTS semester;
