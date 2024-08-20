"use client";
import { useEffect, useState } from "react";
import { ChartConfig, ChartContainer } from "@/components/ui/chart";
import { Bar, BarChart, CartesianGrid, XAxis } from "recharts";
import { ChartTooltip, ChartTooltipContent } from "@/components/ui/chart";
import { ChartLegend, ChartLegendContent } from "@/components/ui/chart";
import { useRouter } from "next/navigation";

type UserProgress = {
  id: number;
  user_id: number;
  book_id: number;
  chapter_chats_completed: number[];
  chat_score_attempted: number;
  chats_achieved: number;
  quizzes_score_attempted: number;
  quizzes_score_achieved: number;
  quiz_chapters_completed: number[];
  total_score_attempted: number;
  total_score_achieved: number;
};

export default function CourseProgress({
  params,
}: {
  params: { bookid: string };
}) {
  const [chapterChatsCompleted, setChapterChatsCompleted] = useState<number[]>(
    [],
  );
  const [chatScoreAttempted, setChatScoreAttempted] = useState<number>();
  const [chatsAchieved, setChatsAchieved] = useState<number>();
  const [quizzesScoreAttempted, setQuizzesScoreAttempted] = useState<number>();
  const [quizzesScoreAchieved, setQuizzesScoreAchieved] = useState<number>();
  const [quizChaptersCompleted, setQuizChaptersCompleted] = useState<number[]>(
    [],
  );
  const [totalScoreAttempted, setTotalScoreAttempted] = useState<number>();
  const [totalScoreAchieved, setTotalScoreAchieved] = useState<number>();
  const [errMsg, setErrMsg] = useState<string>();
  const router = useRouter();

  useEffect(() => {
    fetch(`/api/user/progress/book/${params.bookid}`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
    })
      .then((res) => res.json())
      .then((data) => {
        if (data.error) {
          if (data.error === "invalid or missing authentication token") {
            router.replace("/");
            return;
          }
          console.error(data.error);
          return;
        }
        console.log(data.user_progress);
        setChapterChatsCompleted(data.user_progress.chapter_chats_completed);
        setChatScoreAttempted(data.user_progress.chat_score_attempted);
        setChatsAchieved(data.user_progress.chats_achieved);
        setQuizzesScoreAttempted(data.user_progress.quizzes_score_attempted);
        setQuizzesScoreAchieved(data.user_progress.quizzes_score_achieved);
        setQuizChaptersCompleted(data.user_progress.quiz_chapters_completed);
        setTotalScoreAttempted(data.user_progress.total_score_attempted);
        setTotalScoreAchieved(data.user_progress.total_score_achieved);
      })
      .catch((error) => {
        setErrMsg(error);
      });
  }, []);

  const progressData = [
    {
      category: "Chats",
      attempted: chatScoreAttempted,
      achieved: chatsAchieved,
    },
    {
      category: "Quizzes",
      attempted: quizzesScoreAttempted,
      achieved: quizzesScoreAchieved,
    },
    {
      category: "Totals",
      attempted: totalScoreAttempted,
      achieved: totalScoreAchieved,
    },
  ];

  const chartConfig = {
    attempted: {
      label: "Attempted",
      color: "#2563eb",
    },
    achieved: {
      label: "Achieved",
      color: "#60a5fa",
    },
  } satisfies ChartConfig;

  return (
    <section className="flex flex-col justify-center items-center gap-12 py-12 ">
      <h1 className="font-display text-4xl">My Progress</h1>
      {errMsg && <p className="text-red-600">{errMsg}</p>}
      <ChartContainer
        config={chartConfig}
        className="min-h-[200px] md:w-[600px] w-[380px]"
      >
        <BarChart accessibilityLayer data={progressData}>
          <CartesianGrid vertical={false} />
          <XAxis
            dataKey="category"
            tickLine={false}
            tickMargin={10}
            axisLine={false}
            tickFormatter={(value) => value}
          />
          <ChartTooltip content={<ChartTooltipContent />} />
          <ChartLegend content={<ChartLegendContent />} />
          <Bar dataKey="attempted" fill="var(--color-attempted)" radius={4} />
          <Bar dataKey="achieved" fill="var(--color-achieved)" radius={4} />
        </BarChart>
      </ChartContainer>
    </section>
  );
}
