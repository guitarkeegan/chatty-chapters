"use client";
import TeacherLoginForm from "@/components/TeacherLoginForm";
import Link from "next/link";

export default function TeacherLogin() {
  return (
    <main>
      <section className="bg-main h-screen flex flex-col gap-24 justify-center items-center">
        <div>
          <h2 className="font-display text-4xl text-center">Teacher Login</h2>
        </div>
        <div>
          <TeacherLoginForm />
        </div>
        <Link href="/forgot">Forgot password?</Link>
      </section>
    </main>
  );
}
