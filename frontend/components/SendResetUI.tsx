"use client";
import { useState } from "react";
export default function SendResetUI() {
  const [username, setUsername] = useState("");
  const [errorMsg, setErrorMsg] = useState("");
  const [message, setMessage] = useState("");

  const sendReset = (e: React.FormEvent) => {
    e.preventDefault();

    fetch("api/reset-password", {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        username: username,
      }),
    })
      .then((response) => response.json())
      .then((data) => {
        if (data.error) {
          setErrorMsg(data.error);
          return;
        }
        if (data.status === "SUCCESS") {
          setMessage(data.message);
        }
      })
      .catch((error) => {
        setErrorMsg("Error on Login: " + error);
      });
  };

  return (
    <form
      onSubmit={sendReset}
      className="flex flex-col gap-4 items-center border-black border pt-12 pb-12 px-4 font-body"
    >
      <p className="text-red-600">{errorMsg && errorMsg}</p>
      <p className="text-green-600">{message && message}</p>
      <label htmlFor="username">email</label>
      <input
        type="text"
        className="bg-main border border-black"
        name="username"
        value={username}
        onChange={(e) => setUsername(e.target.value)}
      />
      <button className="px-4 mt-4 border border-black" type="submit">
        Enter
      </button>
    </form>
  );
}
