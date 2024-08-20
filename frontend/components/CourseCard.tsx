import Link from "next/link";
export type CourseCardProps = {
  course_id: number;
  course_title: string;
  semester: string;
  year: number;
  book_title: string;
  book_id: number;
};

export default function CourseCard({
  course_id,
  course_title,
  semester,
  year,
  book_title,
  book_id,
}: CourseCardProps) {
  return (
    <Link
      className="border border-black w-[240px] h-[200px] flex justify-center flex-col
      items-center hover:bg-focus transition-colors duration-200 ease-in-out text-center"
      href={`/dashboard/my-progress/${book_id}`}
    >
      <h2 className="font-body text-2xl">Course: {course_title}</h2>
      <div className="flex flex-col">
        <p className="font-body text-lg">Course ID: {course_id}</p>
        <p className="font-body text-lg">
          {semester[0].toUpperCase() + semester.substring(1)}: {year}
        </p>
        <p className="font-body text-lg">Textbook: {book_title}</p>
      </div>
    </Link>
  );
}
