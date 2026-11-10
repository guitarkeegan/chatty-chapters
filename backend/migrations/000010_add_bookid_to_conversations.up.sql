ALTER TABLE ai_user_conversations
ADD COLUMN book_id BIGINT REFERENCES books(id) ON DELETE CASCADE;
