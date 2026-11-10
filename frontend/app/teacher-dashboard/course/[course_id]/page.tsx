"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import ChapterCard from "@/components/ChapterCard";
import { useSearchParams, useParams } from "next/navigation";

type Chapter = {
  id: number;
  chapter_number: number;
  title: string;
};

export default function TeacherCourse() {
  const router = useRouter();
  const [chapters, setChapters] = useState<Chapter[]>([]);
  const [errMsg, setErrMsg] = useState<string>();
  const searchParams = useSearchParams();
  const bookId = searchParams.get("book_id");
  const { course_id } = useParams();

  useEffect(() => {
    fetch(`/api/teacher/course/book/${bookId}`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
    })
      .then((res) => {
        if (res.status === 401) {
          router.replace("/dashboard");
        }
        return res.json();
      })
      .then((data) => {
        if (data.error) {
          router.replace("/teacher-login");
          return;
        }
        if (data.chapters) {
          setChapters(data.chapters);
        }
      })
      .catch((error) => {
        setErrMsg(error);
      });
  }, []);

  if (!bookId) {
    return <p>Book ID not found</p>;
  }
  return (
    <section className="flex flex-col justify-center items-center gap-12 py-12 ">
      <h1 className="font-display text-4xl">Chapters</h1>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {errMsg ? (
          <p>{errMsg}</p>
        ) : (
          chapters &&
          chapters.length > 0 &&
          chapters.map((chapter, id) => (
            <ChapterCard
              key={id}
              id={chapter.id}
              title={chapter.title}
              chapter_number={chapter.chapter_number}
              bookid={bookId}
              course_id={course_id as string}
            />
          ))
        )}
      </div>
    </section>
  );
}
