CREATE TABLE IF NOT EXISTS quizzes (

    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    book_id BIGINT REFERENCES books(id) ON DELETE CASCADE,
    chapter_id BIGINT REFERENCES chapters(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    attempted BOOLEAN DEFAULT false,
    score INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS quiz_questions (

    id BIGSERIAL PRIMARY KEY,
    quiz_id BIGINT REFERENCES quizzes(id) ON DELETE CASCADE,
    quiz_question TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS quiz_responses (

    id BIGSERIAL PRIMARY KEY,
    quiz_id BIGINT REFERENCES quizzes(id) ON DELETE CASCADE,
    quiz_question_id BIGINT REFERENCES quiz_questions(id) ON DELETE CASCADE,
    student_response TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS quiz_response_evals (

    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    book_id BIGINT REFERENCES books(id) ON DELETE CASCADE,
    chapter_id BIGINT REFERENCES chapters(id) ON DELETE CASCADE,
    quiz_id BIGINT REFERENCES quizzes(id) ON DELETE CASCADE,
    quiz_question_id BIGINT REFERENCES quiz_questions(id) ON DELETE CASCADE,
    quiz_response_id BIGINT REFERENCES quiz_responses(id) ON DELETE CASCADE,
    score INTEGER DEFAULT 0,
    response_feedback TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
