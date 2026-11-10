CREATE TABLE IF NOT EXISTS courses_books (
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    book_id BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    PRIMARY KEY (course_id, book_id)
);

CREATE TABLE IF NOT EXISTS users_courses (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, course_id)
);
