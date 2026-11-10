import { Book } from "@/app/dashboard/page";
import Link from "next/link";

export default function BookCard(props: Book) {
  return (
    <Link
      className="border border-black w-[200px] h-[200px] flex justify-center
      items-center hover:bg-focus transition-colors duration-200 ease-in-out text-center"
      href={`/dashboard/book/${props.id}`}
    >
      <h2 className="font-display text-2xl">{props.title}</h2>
    </Link>
  );
}
