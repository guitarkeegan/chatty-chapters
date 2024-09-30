"use client";
import { useSearchParams, useParams, useRouter } from "next/navigation";
import { useState, useEffect } from "react";
import ResultsTable from "@/components/ResultsTable";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";

export type Result = {
  user_id: number;
  first_name: string;
  last_name: string;
  chat_score: number;
  quiz_score: number;
  total_activity_score: number;
};

type UserChapterResults = {
  questions: string[];
  quiz_questions: string[];
  quiz_answers: string[];
  quiz_feedback: string[];
};

export default function Results() {
  const searchParams = useSearchParams();
  const bookId = searchParams.get("book_id");
  const { chapter_id, course_id } = useParams();
  const [results, setResults] = useState<Result[]>([]);
  const [errMsg, setErrMsg] = useState<string>("");
  const [dialogOpen, setDialogOpen] = useState<boolean>(false);
  const [studentResults, setStudentResults] = useState<UserChapterResults>();
  const router = useRouter();

  useEffect(() => {
    fetch(`/api/teacher/course/chapter/results`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify({
        book_id: Number(bookId),
        chapter_id: Number(chapter_id),
        course_id: Number(course_id),
      }),
    })
      .then((res) => {
        if (res.status == 401) {
          router.replace("/dashboard");
        }
        return res.json();
      })
      .then((data) => {
        if (data.error) {
          setErrMsg(data.error);
          return;
        }
        setResults(data.results);
      })
      .catch((err) => {
        setErrMsg(err.message);
      });
  }, []);

  const handleRowClick = (result: Result) => {
    console.log("row clicked");
    getStudentResults(result);
  };

  const getStudentResults = async (result: Result) => {
    try {
      const res = await fetch(`/api/teacher/course/chapter/results/student`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify({
          user_id: result.user_id,
          book_id: Number(bookId),
          chapter_id: Number(chapter_id),
        }),
      });
      const data = await res.json();
      console.log(data);
      if (data.error === "invalid or missing authentication token") {
        setErrMsg(data.error);
        router.replace("/teacher-login");
        return;
      }
      setStudentResults(data.student_results);
      setDialogOpen(true); // Open the dialog after fetching results
    } catch (error) {
      console.error(error);
    }
  };

  return (
    <>
      {errMsg ? (
        <div>{errMsg}</div>
      ) : (
        <>
          <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Student Results</DialogTitle>
                <DialogDescription>
                  This is the activity recorded for this student.
                </DialogDescription>
              </DialogHeader>
              <ScrollArea className="h-[500px]">
                <div>
                  <h3 className="text-lg font-semibold">Chat Questions</h3>
                  <ul>
                    {studentResults?.questions.map((question, index) => (
                      <li
                        className="py-2 my-1 px-2 bg-main rounded-md"
                        key={index}
                      >
                        {question}
                      </li>
                    ))}
                  </ul>
                </div>
                <div>
                  <h3 className="text-lg font-semibold">Quiz Questions</h3>
                  <ul>
                    {studentResults?.quiz_questions.map((question, index) => (
                      <li
                        className="py-2 my-1 px-2 bg-main rounded-md"
                        key={index}
                      >
                        {question}
                      </li>
                    ))}
                  </ul>
                </div>
                <div>
                  <h3 className="text-lg font-semibold">Quiz Answers</h3>
                  <ul>
                    {studentResults?.quiz_answers.map((answer, index) => (
                      <li
                        className="py-2 my-1 px-2 bg-main rounded-md"
                        key={index}
                      >
                        {answer}
                      </li>
                    ))}
                  </ul>
                </div>
                <div>
                  <h3 className="text-lg font-semibold">Quiz Feedback</h3>
                  <ul>
                    {studentResults?.quiz_feedback.map((feedback, index) => (
                      <li
                        className="py-2 my-1 px-2 bg-focus rounded-md"
                        key={index}
                      >
                        {feedback}
                      </li>
                    ))}
                  </ul>
                </div>
              </ScrollArea>
            </DialogContent>
          </Dialog>
          <ResultsTable
            rowClick={handleRowClick}
            results={results}
          ></ResultsTable>
        </>
      )}
    </>
  );
}
