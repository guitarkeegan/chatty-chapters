import { Chapter } from "@/app/dashboard/book/[bookid]/page";
import Link from "next/link";

interface ChapterCardProps extends Chapter {
  bookid: string;
  course_id?: string;
}

export default function ChapterCard({
  bookid,
  id,
  chapter_number,
  title,
  course_id,
}: ChapterCardProps) {
  return (
    <>
      {course_id ? (
        <Link
          className="border border-black w-[200px] h-[120px] flex justify-start
      items-start hover:bg-focus transition-colors duration-200 ease-in-out text-ellipsis overflow-hidden"
          href={`/teacher-dashboard/course/${course_id}/chapter/${id}/results?book_id=${bookid}`}
        >
          <div className="flex flex-col justify-start items-start p-4">
            <h2 className="font-display text-xl">Chapter {chapter_number}:</h2>
            <p className="font-body text-md">{title}</p>
          </div>
        </Link>
      ) : (
        <Link
          className="border border-black w-[200px] h-[120px] flex justify-start
      items-start hover:bg-focus transition-colors duration-200 ease-in-out text-ellipsis overflow-hidden"
          href={`/dashboard/book/${bookid}/chapter/${id}/activities`}
        >
          <div className="flex flex-col justify-start items-start p-4">
            <h2 className="font-display text-xl">Chapter {chapter_number}:</h2>
            <p className="font-body text-md">{title}</p>
          </div>
        </Link>
      )}
    </>
  );
}
