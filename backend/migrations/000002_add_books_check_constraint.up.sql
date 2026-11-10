ALTER TABLE users ADD CONSTRAINT books_length_check CHECK (array_length(books, 1) > 0);
