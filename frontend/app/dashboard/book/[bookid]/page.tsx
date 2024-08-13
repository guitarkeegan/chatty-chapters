"use client";
import ChapterCard from "@/components/ChapterCard";
import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";

export type Chapter = {
  id: number;
  chapter_number: number;
  title: string;
};

export default function Book({ params }: { params: { bookid: string } }) {
  const [chapters, setChapters] = useState<Chapter[]>([]);
  const [errMsg, setErrMsg] = useState<string>();

  const router = useRouter();

  useEffect(() => {
    fetch(`/api/chapters/book/${params.bookid}`, {
      method: "GET",
      credentials: "include",
      headers: {
        Accept: "application/json",
        "Content-type": "application/json",
      },
    })
      .then((response) => response.json())
      .then((data) => {
        if (data.error) {
          console.error(data.error);
          router.replace("/");
          return;
        }
        setChapters(data.chapters);
      })
      .catch((error) => setErrMsg(`Error getting books: ${error}`));
  }, []);

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
              bookid={params.bookid}
            />
          ))
        )}
      </div>
    </section>
  );
}
