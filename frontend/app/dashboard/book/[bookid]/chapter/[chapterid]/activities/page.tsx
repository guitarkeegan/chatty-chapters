"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import ActivityCard from "@/components/ActivityCard";
// SELECT id, user_id, book_id, chapter_id, activity_type, score, completed_at, attempted

export type Activies = {
  id: number;
  user_id: number;
  book_id: number;
  chapter_id: number;
  activity_type: string;
  score: number;
  completed_at: Date;
  attempted: boolean;
};

export default function Activities({
  params,
}: {
  params: { bookid: number; chapterid: number };
}) {
  const [activities, setActivities] = useState<Activies[]>([]);
  const [errMsg, setErrMsg] = useState<string>();

  const router = useRouter();

  // react query?
  // https://tanstack.com/query/latest
  useEffect(() => {
    fetch(`/api/activities/book/${params.bookid}/chapter/${params.chapterid}`, {
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
        setActivities(data.activities);
      })
      .catch((error) => setErrMsg(`Error getting activities: ${error}`));
  }, []);

  return (
    <section className="flex flex-col justify-center items-center gap-12 py-12 ">
      <h1 className="font-display text-4xl">Activites</h1>
      {errMsg ? (
        <p>{errMsg}</p>
      ) : (
        activities.length > 0 &&
        activities.map((activity, id) => (
          <ActivityCard
            key={id}
            activity_type={activity.activity_type}
            attempted={activity.attempted}
            score={activity.score}
            id={id}
            user_id={activity.user_id}
            book_id={activity.book_id}
            chapter_id={activity.chapter_id}
            completed_at={activity.completed_at}
          />
        ))
      )}
    </section>
  );
}
