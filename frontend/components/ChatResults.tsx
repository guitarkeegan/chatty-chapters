import { type ChapterChatHistory, ChapterChatEval } from "@/app/dashboard/book/[bookid]/chapter/[chapterid]/activities/chat/page";

type ChatResultsProps = {
    history: ChapterChatHistory[];
    evaluation: ChapterChatEval | undefined;
    update: Function;
}

export default function ChatResults({history, evaluation, update}: ChatResultsProps) {

    console.log(history);
    console.log(evaluation);

    return (
        <section className="flex flex-col items-center gap-2 py-12">
            <h1 className="font-display text-4xl py-12">Chapter Chat History and Evaluation</h1>
            {
                history.length > 0 &&
                (
                    history.map((msgs, idx) => (
                        <div className="sm:max-w-[600px] p-4 bg-focus" key={idx}>
                            <p><span className="font-body font-bold">You: </span>{msgs.user_message}</p>
                            <p><span className="font-body font-bold">AI: </span>{msgs.ai_message}</p>
                        </div>
                    ))
                )
            }
            {
                evaluation && (
                    <div className="sm:max-w-[600px] bg-attempted text-white p-4">
                        <p><span className="font-body font-bold">AI Feedback: </span>{evaluation.feedback}</p>
                        <p><span className="font-body font-bold">Score: </span>{evaluation.score}</p>
                    </div>
                )
            }
            <button className="border border-black p-4" onClick={() => update()}>Continue to chat</button>
        </section>
    )
}
