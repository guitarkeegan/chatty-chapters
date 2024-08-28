"use client";
import ChatResults from "@/components/ChatResults";
import ChatUI from "@/components/ChatUI";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

export type ChapterChatHistory = {
  user_message: string;
  user_message_time: Date;
  ai_message: string;
  ai_mesage_time: Date;
};

export type ChapterChatEval = {
  id: number;
  created_at: Date;
  score: number;
  feedback: string;
};

export default function Chat({
  params,
}: {
  params: { bookid: number; chapterid: number };
}) {
  const [errMsg, setErrMsg] = useState<string>("");
  const [chapterChatHistory, setChapterChatHistory] = useState<
    ChapterChatHistory[]
  >([]);
  const [chapterChatEval, setChapterChatEval] = useState<ChapterChatEval>();
  const [showResults, setShowResults] = useState<boolean>(true);

  const router = useRouter();

  useEffect(() => {
    fetch(
      `/api/chapters/chat/start/book/${params.bookid}/chapter/${params.chapterid}`,
      {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
      },
    )
      .then((res) => res.json())
      .then((data) => {
        if (data.error === "") {
          if (data.error === "invalid or missing authentication token") {
            router.replace("/");
            return;
          }
        }
        console.log(data.chat.history);
        console.log(data.chat.eval);
        if (!data.chat.eval || !data.chat.history) {
          console.log("not eval or history");
          setShowResults(false);
          return;
        }
        setShowResults(true);
        setChapterChatHistory(data.chat.history);
        setChapterChatEval(data.chat.eval);
      })
      .catch((err) => setErrMsg(err));
  }, []);

  const updateShowResults = () => {
    setShowResults(false);
  };

  return (
    <>
      {showResults ? (
        <ChatResults
          history={chapterChatHistory}
          evaluation={chapterChatEval}
          update={updateShowResults}
        />
      ) : (
        <ChatUI bookid={params.bookid} chapterid={params.chapterid} />
      )}
    </>
  );
}
