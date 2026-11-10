ALTER TABLE user_messages DROP CONSTRAINT IF EXISTS fk_book_id;
ALTER TABLE user_messages DROP COLUMN IF EXISTS book_id;

ALTER TABLE user_messages DROP CONSTRAINT IF EXISTS fk_chapter_id;
ALTER TABLE user_messages DROP COLUMN IF EXISTS chapter_id;
