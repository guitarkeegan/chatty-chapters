"use client";
import TeacherLoginForm from "@/components/TeacherLoginForm";
import Link from "next/link";
import { useEffect } from "react";
import { useRouter } from "next/navigation";

export default function TeacherLogin() {
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
        if (data.logged_in) {
          router.replace("/teacher-dashboard");
        }
      })
      .catch((error) => {
        console.error(error);
      });
  }, []);

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
