"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import BookCard from "@/components/BookCard";

export type Book = {
  id: number;
  title: string;
};

export default function Dashboard() {
  const [books, setBooks] = useState<Book[]>([]);
  const [errMsg, setErrMsg] = useState<string>();

  const router = useRouter();

  useEffect(() => {
    fetch("/api/books", {
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
        setBooks(data.books);
      })
      .catch((error) => setErrMsg(`Error getting books: ${error}`));
  }, []);

  return (
    <section className="flex flex-col justify-center items-center gap-12 py-12 ">
      <h1 className="font-display text-4xl">Books</h1>
      {errMsg ? (
        <p>{errMsg}</p>
      ) : (
        books.length > 0 &&
        books.map((book, id) => (
          <BookCard key={id} id={book.id} title={book.title} />
        ))
      )}
    </section>
  );
}
