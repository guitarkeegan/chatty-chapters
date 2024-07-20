import { Activies } from "@/app/dashboard/book/[bookid]/chapter/[chapterid]/activities/page";
import Link from "next/link";

export default function ActivityCard(props: Activies) {
  return (
    <Link
      className="border border-black w-[200px] h-[120px] flex justify-center
      items-center hover:bg-focus transition-colors duration-200 ease-in-out"
      href={`/dashboard/book/${props.book_id}/chapter/${props.chapter_id}/activities/${props.activity_type}`}
    >
      <div className="flex flex-col items-start p-4">
        <h2 className="font-display text-xl">{props.activity_type}</h2>
        <p className="font-body text-md">
          {props.attempted ? props.score : "Start"}
        </p>
      </div>
    </Link>
  );
}
