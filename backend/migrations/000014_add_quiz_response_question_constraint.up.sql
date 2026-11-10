ALTER TABLE quiz_responses
ADD CONSTRAINT unique_quiz_question_response UNIQUE (quiz_id, quiz_question_id);
