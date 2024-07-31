"use client";
import { Suspense } from "react";
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
    question_id: number;
    student_response: string;
  };

  const [questions, setQuestions] = useState<QuizQuestion[]>([]);
  const [responses, setResponses] = useState<QuizResponse[]>([]);
  const [inputText, setInputText] = useState("");
  const [currentQuestionIdx, setCurrentQuestionIdx] = useState(0);
  const [errorMsg, setErrorMsg] = useState("");

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
        setQuestions(data.quiz);
      })
      .catch((error) => {
        // TODO: redirect to home
        console.error(error);
      });
  }, []);

  const send = () => {
    console.log(inputText);

    if (questions && questions.length > 0) {
      // store response and go to next question
      const response: QuizResponse = {
        quiz_id: questions[currentQuestionIdx].quiz_id,
        question_id: questions[currentQuestionIdx].id,
        student_response: questions[currentQuestionIdx].quiz_question,
      };
      setResponses((prev) => [...prev, response]);
      setCurrentQuestionIdx(currentQuestionIdx + 1);
    } else {
      setErrorMsg("no quiz questions, tell Keegan he is lame");
    }
    // Check for end of quiz
    // send to api
    if (currentQuestionIdx === questions.length) {
      console.log("quiz completed!");
      // TODO: akually send the stuff
      setCurrentQuestionIdx(currentQuestionIdx + 1);
    }

    setInputText("");
  };

  const handleKeyPress = (e: any) => {
    if (e.key === "Enter") {
      e.preventDefault();
      send();
    }
  };

  const handleInputChange = (e: any) => {
    setInputText(e.target.value);
  };

  return (
    <section className="flex flex-col justify-between items-center">
      <div className="h-[410px] flex flex-col gap-y-16 items-center">
        <h1 className="font-display text-4xl">Quiz</h1>
        {errorMsg && <div className="text-red-600">{errorMsg}</div>}
        <div className="flex gap-4 justify-center">
          <div className="w-[10px] h-[10px] bg-black" />
          <div
            style={
              currentQuestionIdx >= 1
                ? { backgroundColor: "black" }
                : { backgroundColor: "white" }
            }
            className="w-[10px] h-[10px] transition-colors ease-in-out"
          />
          <div
            style={
              currentQuestionIdx === 2
                ? { backgroundColor: "black" }
                : { backgroundColor: "white" }
            }
            className="w-[10px] h-[10px] transition-colors ease-in-out"
          />
        </div>
        <div className="font-body max-w-[260px] transition-all ease-in-out">
          {questions.length > 0
            ? currentQuestionIdx > 2
              ? "Quiz Completed!"
              : questions[currentQuestionIdx].quiz_question
            : "no questions"}
        </div>
      </div>
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
    </section>
  );
}
