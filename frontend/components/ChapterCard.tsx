import { Chapter } from "@/app/dashboard/book/[bookid]/page";
import Link from "next/link";

interface ChapterCardProps extends Chapter {
  bookid: string;
}

export default function ChapterCard(props: ChapterCardProps) {
  return (
    <Link
      className="border border-black w-[200px] h-[120px] flex justify-start
      items-start hover:bg-focus transition-colors duration-200 ease-in-out text-ellipsis overflow-hidden"
      href={`/dashboard/book/${props.bookid}/chapter/${props.id}/activities`}
    >
      <div className="flex flex-col justify-start items-start p-4">
        <h2 className="font-display text-xl">
          Chapter {props.chapter_number}:
        </h2>
        <p className="font-body text-md">{props.title}</p>
      </div>
    </Link>
  );
}
