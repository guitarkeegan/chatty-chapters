"use client";
import LoginForm from "@/components/LoginForm";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import SignupForm from "@/components/SignUpForm";

export default function Home() {
  const [signUp, setSignUp] = useState<boolean>(false);

  const getSignupForm = (b: boolean) => {
    setSignUp(b);
  };

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
          router.replace("/dashboard");
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
          <h2 className="font-display text-4xl text-center">Music Tutor</h2>
        </div>
        <div>
          {!signUp ? (
            <LoginForm signup={getSignupForm} />
          ) : (
            <SignupForm login={getSignupForm} />
          )}
        </div>
      </section>
    </main>
  );
}
