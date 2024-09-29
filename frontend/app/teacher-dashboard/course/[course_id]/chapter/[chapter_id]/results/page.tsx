"use client";
import { useSearchParams, useParams, useRouter } from "next/navigation";
import { useState, useEffect } from "react";
import ResultsTable from "@/components/ResultsTable";

export type Result = {
  user_id: number;
  first_name: string;
  last_name: string;
  chat_score: number;
  quiz_score: number;
  total_activity_score: number;
};

export default function Results() {
  const searchParams = useSearchParams();
  const bookId = searchParams.get("book_id");
  const { chapter_id, course_id } = useParams();
  const [results, setResults] = useState<Result[]>([]);
  const [errMsg, setErrMsg] = useState<string>("");
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

  return (
    <>
      {errMsg ? (
        <div>{errMsg}</div>
      ) : (
        <ResultsTable results={results}></ResultsTable>
      )}
    </>
  );
}
