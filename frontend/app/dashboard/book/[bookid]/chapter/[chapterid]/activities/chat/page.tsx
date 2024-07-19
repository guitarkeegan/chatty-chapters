import ChatUI from "@/components/ChatUI";

export default function Chat({
  params,
}: {
  params: { bookid: number; chapterid: number };
}) {
  return <ChatUI bookid={params.bookid} chapterid={params.chapterid} />;
}
