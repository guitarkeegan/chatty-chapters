CREATE TABLE IF NOT EXISTS user_progress (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    book_id BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    chapter_chats_completed INTEGER[] DEFAULT '{}',
    chat_score_attempted INTEGER NOT NULL DEFAULT 0,
    chats_achieved INTEGER NOT NULL DEFAULT 0,
    quizzes_score_attempted INTEGER NOT NULL DEFAULT 0,
    quizzes_score_achieved INTEGER NOT NULL DEFAULT 0,
    quiz_chapters_completed INTEGER[] DEFAULT '{}',
    total_score_attempted INTEGER NOT NULL DEFAULT 0,
    total_score_achieved INTEGER NOT NULL DEFAULT 0
);
