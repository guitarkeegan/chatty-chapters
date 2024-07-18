"use client";
import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";

export default function LoginForm() {
  const [email, setEmail] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [errorMsg, setErrorMsg] = useState<string>("");

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

  const router = useRouter();

  const login = (e: React.FormEvent) => {
    e.preventDefault();
    fetch("api/login", {
      method: "POST",
      credentials: "include",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        username: email,
        password,
      }),
    })
      .then((response) => response.json())
      .then((data) => {
        if (data.error) {
          setErrorMsg(data.error);
          return;
        }
        console.log(data);
        router.replace("dashboard");
      })
      .catch((error) => {
        setErrorMsg("Error on Login: " + error);
      });
  };

  return (
    <form
      onSubmit={login}
      className="flex flex-col gap-4 items-center border-black border pt-20 pb-12 px-4 font-body"
    >
      <p className="text-red-600">{errorMsg && errorMsg}</p>
      <label htmlFor="email">Email</label>
      <input
        className="bg-main border border-black"
        name="email"
        type="email"
        onChange={(e) => setEmail(e.target.value)}
      />
      <label htmlFor="password">Password</label>
      <input
        className="bg-main border border-black"
        name="password"
        type="password"
        onChange={(e) => setPassword(e.target.value)}
      />
      <button className="px-4 mt-4 border border-black" type="submit">
        Enter
      </button>
    </form>
  );
}
