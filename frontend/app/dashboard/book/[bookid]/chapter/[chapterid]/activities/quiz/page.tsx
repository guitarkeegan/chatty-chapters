import QuizUI from "@/components/QuizUI";
export default function Quiz({
  params,
}: {
  params: { bookid: number; chapterid: number };
}) {
  return <QuizUI bookid={params.bookid} chapterid={params.chapterid} />;
}
