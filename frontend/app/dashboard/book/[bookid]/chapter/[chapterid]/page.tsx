import ChatUI from "@/components/ChatUI";

export default function Chapters({
  params,
}: {
  params: { bookid: number; chapterid: number };
}) {
  return <ChatUI bookid={params.bookid} chapterid={params.chapterid} />;
}
