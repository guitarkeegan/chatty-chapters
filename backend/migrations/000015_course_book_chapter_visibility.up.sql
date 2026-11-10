CREATE TABLE IF NOT EXISTS course_book_chapter_visibility (
    book_id BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    chapter_id BIGINT NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    visible BOOLEAN NOT NULL DEFAULT true,
    PRIMARY KEY (book_id, chapter_id, course_id)
);
