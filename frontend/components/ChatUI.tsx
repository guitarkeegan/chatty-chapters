"use client";
import { useRouter } from "next/navigation";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { HiOutlinePaperAirplane } from "react-icons/hi2";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { ScrollArea } from "./ui/scroll-area";
import { useState, useEffect, useRef } from "react";

type Participant = {
  role: string;
  text: string;
};

export default function ChatUI({
  bookid,
  chapterid,
}: {
  bookid: number;
  chapterid: number;
}) {
  const [inputText, setInputText] = useState<Participant>({
    role: "user",
    text: "",
  });
  const [outputTexts, setOutputTexts] = useState<Participant[]>([]);
  const [currentAIResponse, setCurrentAIResponse] = useState<string>("");
  const websocketRef = useRef<WebSocket | null>(null);
  const [aiGenerating, setAiGenerating] = useState(false);
  const [aiStart, setAiStart] = useState(false);

  const router = useRouter();

  useEffect(() => {
    fetch("/api/user/logged-in", {
      method: "GET",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      credentials: "include",
    })
      .then((response) => response.json())
      .then((data) => {
        if (data.error) {
          if (data.error === "invalid or missing authentication token") {
            router.replace("/");
            return;
          }
          console.error(data.error);
          return;
        }
      })
      .catch((error) => {
        console.error(error);
      });
  }, []);

  useEffect(() => {
    const wsUrl = `${process.env.NEXT_PUBLIC_WS_BASE_URL}/${bookid}/chapter/${chapterid}`;
    const ws = new WebSocket(wsUrl);
    websocketRef.current = ws;

    ws.onopen = () => {
      console.log("WebSocket connection established");
    };

    ws.onmessage = (event) => {
      setAiGenerating(true);
      if (currentAIResponse.length === 0) {
        setAiStart(true);
      } else {
        setAiStart(false);
      }
      if (event.data === "\n\r\n\r") {
        setAiGenerating(false);
      }
      const message = event.data;
      setCurrentAIResponse((prev) => prev + message);
    };

    ws.onclose = () => {
      console.log("WebSocket connection closed");
    };

    ws.onerror = (error) => {
      console.error("WebSocket error", error);
    };

    // Cleanup on unmount
    return () => {
      if (ws) {
        ws.close();
      }
    };
  }, []);

  useEffect(() => {
    if (currentAIResponse) {
      // Update the outputTexts with the current streaming response
      setOutputTexts((prev) => {
        const lastMessage = prev[prev.length - 1];
        if (lastMessage && lastMessage.role === "ai") {
          return [
            ...prev.slice(0, -1),
            { ...lastMessage, text: currentAIResponse },
          ];
        } else {
          return [...prev, { role: "ai", text: currentAIResponse }];
        }
      });
    }
  }, [currentAIResponse]);

  const send = () => {
    if (websocketRef.current && inputText.text.trim() !== "") {
      websocketRef.current.send(inputText.text);
      setOutputTexts((prev) => [...prev, inputText]);
      setInputText({ role: "user", text: "" });
      setCurrentAIResponse(""); // Reset current AI response
    }
  };

  const handleKeyPress = (e: any) => {
    if (e.key === "Enter") {
      e.preventDefault();
      send();
    }
  };

  const handleInputChange = (e: any) => {
    setInputText({ role: "user", text: e.target.value });
  };

  return (
    <section className="flex flex-col px-2 justify-between items-center">
      <ScrollArea className="h-[410px] lg:h-[500px] text-lg font-body max-w-[600px]">
        {outputTexts ? (
          outputTexts.map((participant, i) =>
            participant.role === "ai" && aiStart ? (
              <div key={i} className="ai-bubble">
                <Avatar>
                  <AvatarImage src="https://github.com/shadcn.png" />
                  <AvatarFallback>CN</AvatarFallback>
                </Avatar>
                <div className="ai">{participant.text}</div>
              </div>
            ) : participant.role === "ai" ? (
              <div className="ai">{participant.text}</div>
            ) : (
              <div key={i} className="user-bubble flex flex-col items-end">
                <Avatar>
                  {/* TODO: figure out user avitar situation */}
                  <AvatarImage src="https://github.com/shadcn.png" />
                  <AvatarFallback>CN</AvatarFallback>
                </Avatar>
                <div className="user font-body">{participant.text}</div>
              </div>
            ),
          )
        ) : (
          <div>Start chat with AI assistant</div>
        )}
      </ScrollArea>
      <div className="flex flex-row items-end max-w-[600px] w-full font-body">
        <Textarea
          className="font-body"
          id="textarea"
          onChange={handleInputChange}
          value={inputText.text}
          onKeyUp={handleKeyPress}
          placeholder="Type your message here."
        />
        <div className="px-1" />
        <Button onClick={send} size="icon" className="">
          <HiOutlinePaperAirplane className="h-4 w-4" />
        </Button>
      </div>
    </section>
  );
}
