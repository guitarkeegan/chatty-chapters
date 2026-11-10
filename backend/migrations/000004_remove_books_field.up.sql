ALTER TABLE users
DROP COLUMN books;

CREATE TABLE IF NOT EXISTS user_books (
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    book_id BIGINT REFERENCES books(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, book_id)
);
