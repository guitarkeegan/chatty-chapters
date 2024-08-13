"use client";
import { useState } from "react";
import Link from "next/link";

type SignupFormProps = {
  login: Function;
};

export default function SignupForm({ login }: SignupFormProps) {
  const [email, setEmail] = useState<string>("");
  const [firstName, setFirstName] = useState<string>("");
  const [lastName, setLastName] = useState<string>("");
  const [errorMsg, setErrorMsg] = useState<string>("");
  const [msg, setMsg] = useState<string>("");
  const [signedUp, setSignedUp] = useState<boolean>(false);

  const signup = (e: React.FormEvent) => {
    e.preventDefault();
    fetch("api/sign-up", {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        firstName,
        lastName,
        email,
      }),
    })
      .then((response) => response.json())
      .then((data) => {
        if (data.register.status === "SUCCESS") {
          setMsg(data.register.message);
          setSignedUp(true);
        }
      })
      .catch((error) => {
        setErrorMsg("Error on Login: " + error);
      });

    setEmail("");
    setFirstName("");
    setLastName("");
  };

  return (
    <>
      {!signedUp ? (
        <form
          onSubmit={signup}
          className="flex flex-col gap-4 items-center border-black border pt-20 pb-12 px-4 font-body"
        >
          <p className="text-red-600">{errorMsg && errorMsg}</p>
          <label htmlFor="first-name">First</label>
          <input
            className="bg-main border border-black"
            name="first-name"
            type="text"
            value={firstName}
            onChange={(e) => setFirstName(e.target.value)}
          />
          <label htmlFor="last-name">Last</label>
          <input
            className="bg-main border border-black"
            name="last-name"
            type="text"
            value={lastName}
            onChange={(e) => setLastName(e.target.value)}
          />
          <label htmlFor="email">Email</label>
          <input
            className="bg-main border border-black"
            name="email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
          <button className="px-4 mt-4 border border-black" type="submit">
            Sign-up
          </button>
          <button
            className="px-4 mt-4 border border-black"
            type="submit"
            onClick={() => login(false)}
          >
            Login
          </button>
        </form>
      ) : (
        <div className="flex flex-col gap-4 items-center border-black border pt-20 pb-12 px-4 font-body">
          <p className="text-green-800">{msg && msg}</p>
          <button
            className="px-4 mt-4 border border-black"
            type="submit"
            onClick={() => login(false)}
          >
            Sign-up
          </button>
        </div>
      )}
    </>
  );
}
