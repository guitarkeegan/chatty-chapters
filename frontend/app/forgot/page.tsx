"use client";
import ForgotPasswordUI from "@/components/ForgotPasswordUI";
import SendResetUI from "@/components/SendResetUI";
import { useState } from "react";
import Link from "next/link";
export default function Forgot() {
  const [reset, setReset] = useState(true);

  return (
    <section className="bg-main h-screen flex flex-col gap-24 justify-center items-center">
      <div>
        <h2 className="font-display text-4xl text-center">Forgot Password</h2>
        {reset ? <SendResetUI /> : <ForgotPasswordUI />}
      </div>
      <button onClick={() => setReset(false)}>Input confirmation code</button>
      <Link href="/">back to login</Link>
    </section>
  );
}
