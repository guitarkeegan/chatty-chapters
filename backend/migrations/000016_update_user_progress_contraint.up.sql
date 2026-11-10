ALTER TABLE user_progress
ADD CONSTRAINT user_book_unique UNIQUE (user_id, book_id);
