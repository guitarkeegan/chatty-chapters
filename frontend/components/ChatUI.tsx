"use client";
import { HiOutlinePaperAirplane } from "react-icons/hi2";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { ScrollArea } from "./ui/scroll-area";
import { useState } from "react";

export default function ChatUI({
  bookid,
  chapterid,
}: {
  bookid: number;
  chapterid: number;
}) {
  const [inputText, setInputText] = useState<string>("");
  const [outputTexts, setOutputTexts] = useState("");

  const send = () => {
    setOutputTexts(inputText);
    setInputText("");
  };

  const handleKeyPress = (e: any) => {
    if (e.key === "Enter") {
      e.preventDefault();
      send();
    }
  };

  const handleInputChange = (e: any) => {
    setInputText(e.target.value);
  };

  return (
    <section className="flex flex-col px-2 justify-between">
      <ScrollArea className="h-[400px] lg:h-[500px]">{outputTexts}</ScrollArea>
      <div className="flex flex-row items-end justify-between w-full">
        <Textarea
          id="textarea"
          onChange={handleInputChange}
          value={inputText}
          onKeyPress={handleKeyPress}
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
