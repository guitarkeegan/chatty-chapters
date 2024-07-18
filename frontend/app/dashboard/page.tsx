"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

type Book = {
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
          // router.replace("/");
          return;
        }
        setBooks(data.books);
      })
      .catch((error) => setErrMsg(`Error getting books: ${error}`));
  }, []);

  return (
    <section>
      <h1>Books</h1>
      {errMsg ? (
        <p>{errMsg}</p>
      ) : (
        books.length > 0 &&
        books.map((book, id) => (
          <div key={id}>
            <p>{book.id}</p>
            <p>{book.title}</p>
          </div>
        ))
      )}
    </section>
  );
}
