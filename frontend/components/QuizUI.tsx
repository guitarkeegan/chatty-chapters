"use client";
import { HiOutlinePaperAirplane } from "react-icons/hi2";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { useState, useEffect } from "react";

export default function QuizUI({
  bookid,
  chapterid,
}: {
  bookid: number;
  chapterid: number;
}) {
  type QuizQuestion = {
    id: number;
    quiz_id: number;
    quiz_question: string;
    created_at: Date;
  };

  type QuizResponse = {
    quiz_id: number;
    quiz_question_id: number;
    student_response: string;
  };

  type QuizFeedback = {
    quiz_question: string;
    quiz_response: string;
    response_feedback: string;
    score: number;
  };

  const [questions, setQuestions] = useState<QuizQuestion[]>([]);
  const [responses, setResponses] = useState<QuizResponse[]>([]);
  const [inputText, setInputText] = useState("");
  const [currentQuestionIdx, setCurrentQuestionIdx] = useState(0);
  const [errorMsg, setErrorMsg] = useState("");
  const [isQuizCompleted, setIsQuizComplete] = useState(false);
  const [loading, setLoading] = useState<boolean>(true);

  /**
   * quiz result state
   */
  const [quizFeedback, setQuizFeedback] = useState<QuizFeedback[]>([]);

  useEffect(() => {
    fetch(`/api/quiz/take/book/${bookid}/chapter/${chapterid}`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      credentials: "include",
    })
      .then((response) => response.json())
      .then((data) => {
        if (data.error) {
          console.error(data.error);
        }
        if (data.status === "completed") {
          console.log("quiz is done");
          console.log(data.quiz);
          localStorage.removeItem(`quiz${data.quiz[0].quiz_id}`);
          setQuizFeedback(data.quiz);
          setLoading(false);
          return;
        }
        if (data.status === "incomplete") {
          console.log(data.message);
          setErrorMsg(data.message);
        }

        setLoading(false);
        setQuestions(data.quiz);
        const quizId = data.quiz[0].quiz_id;
        // get local storage
        const storedResponses = localStorage.getItem(`quiz${quizId}`);
        if (storedResponses) {
          setResponses(JSON.parse(storedResponses));
          setInputText(JSON.parse(storedResponses)[0].student_response);
        } else {
          // Initialize empty array in local storage if not exists
          localStorage.setItem(`quiz${quizId}`, JSON.stringify([]));
        }
      })
      .catch((error) => {
        // TODO: redirect to home
        console.error(error);
      });
  }, []);

  useEffect(() => {
    if (isQuizCompleted) {
      submitQuiz(responses);
    }
  }, [isQuizCompleted, responses]);

  const send = () => {
    if (questions && questions.length > 0) {
      const quizId = questions[currentQuestionIdx].quiz_id;
      const response: QuizResponse = {
        quiz_id: quizId,
        quiz_question_id: questions[currentQuestionIdx].id,
        student_response: inputText.trim(),
      };

      setResponses((prev) => {
        const existingResponses = [...prev];
        const responseIndex = existingResponses.findIndex(
          (res) => res.quiz_question_id === response.quiz_question_id,
        );

        // Update existing response if found, otherwise add a new one
        if (responseIndex !== -1) {
          existingResponses[responseIndex] = response;
        } else {
          existingResponses.push(response);
        }

        localStorage.setItem(
          `quiz${quizId}`,
          JSON.stringify(existingResponses),
        );
        return existingResponses;
      });

      setCurrentQuestionIdx((prevIdx) => prevIdx + 1);

      // Check for end of quiz
      if (currentQuestionIdx === questions.length - 1) {
        console.log("quiz completed!");
        setIsQuizComplete(true);
        return;
      }

      const nextQuestion = localStorage.getItem(`quiz${quizId}`);
      if (
        nextQuestion &&
        JSON.parse(nextQuestion).length > currentQuestionIdx + 1
      ) {
        setInputText(
          JSON.parse(nextQuestion)[currentQuestionIdx + 1].student_response,
        );
        return;
      }

      setInputText("");
    } else {
      setErrorMsg("no quiz questions, tell Keegan he is lame");
    }
  };

  const submitQuiz = (responses: QuizResponse[]) => {
    console.log("submitting quiz...");
    fetch(`/api/quiz/submit-answers/book/${bookid}/chapter/${chapterid}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      body: JSON.stringify(responses),
      credentials: "include",
    })
      .then((res) => res.json())
      .then((data) => console.log(data))
      .catch((error) => console.error(error));
  };

  const handleKeyPress = (e: any) => {
    if (e.key === "Enter") {
      e.preventDefault();
      if (inputText) {
        send();
      }
    }
  };

  const handleInputChange = (e: any) => {
    setInputText(e.target.value);
  };

  return (
    <section className="flex flex-col justify-between items-center">
      {quizFeedback.length > 0 ? (
        quizFeedback.map((quizFBack, idx) => (
          <div
            key={idx}
            className="flex flex-col gap-y-2 sm:max-w-[600px] max-w-[300px] items-start border border-solid bg-focus p-4"
          >
            <p>
              <span className="font-bold font-body">AI Question:</span>{" "}
              {quizFBack.quiz_question}
            </p>
            <p>
              <span className="font-bold font-body">Student Response:</span>{" "}
              {quizFBack.quiz_response}
            </p>
            <p>
              <span className="font-bold font-body">AI Feedback:</span>{" "}
              {quizFBack.response_feedback}
            </p>
            <p>
              <span className="font-bold font-body">Quiz Score:</span>{" "}
              {quizFBack.score}
            </p>
          </div>
        ))
      ) : (
        <>
          <div className="h-[410px] flex flex-col gap-y-16 items-center">
            <h1 className="font-display text-4xl">Quiz</h1>
            {errorMsg && <div className="text-red-600">{errorMsg}</div>}
            <div className="flex gap-4 justify-center">
              <div className="w-[10px] h-[10px] bg-black" />
              <div
                style={
                  currentQuestionIdx > 0
                    ? { backgroundColor: "black" }
                    : { backgroundColor: "white" }
                }
                className="w-[10px] h-[10px] transition-colors ease-in-out"
              />
              <div
                style={
                  currentQuestionIdx > 1
                    ? { backgroundColor: "black" }
                    : { backgroundColor: "white" }
                }
                className="w-[10px] h-[10px] transition-colors ease-in-out"
              />
            </div>
            <div className="font-body max-w-[260px] transition-all ease-in-out">
              {questions && questions.length > 0
                ? currentQuestionIdx > 2
                  ? "Quiz Completed!"
                  : questions[currentQuestionIdx].quiz_question
                : loading
                  ? "loading"
                  : "no questions"}
            </div>
          </div>
          {currentQuestionIdx <= 2 && !errorMsg && (
            <div className="flex flex-row items-end max-w-[600px] w-full font-body px-2">
              <Textarea
                className="font-body"
                id="textarea"
                onChange={handleInputChange}
                value={inputText}
                onKeyUp={handleKeyPress}
                placeholder="Type your message here."
              />
              <div className="px-1" />
              <Button onClick={send} size="icon" className="">
                <HiOutlinePaperAirplane className="h-4 w-4" />
              </Button>
            </div>
          )}
        </>
      )}
    </section>
  );
}
