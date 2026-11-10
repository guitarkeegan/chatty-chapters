import { Activies } from "@/app/dashboard/book/[bookid]/chapter/[chapterid]/activities/page";
import Link from "next/link";

export default function ActivityCard(props: Activies) {
  const { activity_type, score, book_id, chapter_id } = props;
  const isQuizCompleted = activity_type === "quiz" && score > 0;

  return (
    <>
      {/* {isQuizCompleted ? (
        <div className="border border-black w-[200px] h-[120px] flex justify-center items-center bg-gray-200 cursor-not-allowed">
          <div className="flex flex-col items-start p-4">
            <h2 className="font-display text-xl">{activity_type}</h2>
            <p className="font-body text-md">Completed</p>
          </div>
        </div>
      ) : ( */}
      <Link
        className="border border-black w-[200px] h-[120px] flex justify-center items-center hover:bg-focus transition-colors duration-200 ease-in-out"
        href={`/dashboard/book/${book_id}/chapter/${chapter_id}/activities/${activity_type}`}
      >
        <div className="flex flex-col items-start p-4">
          <h2 className="font-display text-xl">{activity_type}</h2>
          <p className="font-body text-md">
            {score > 0 ? "View Results" : "Start"}
          </p>
        </div>
      </Link>
      {/* )} */}
    </>
  );
}
