"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
export default function ForgotPasswordUI() {
  const router = useRouter();

  const [username, setUsername] = useState("");
  const [confirmationCode, setConfirmationCode] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [errorMsg, setErrorMsg] = useState("");

  const resetPasswordHandler = (e: React.FormEvent) => {
    e.preventDefault();

    fetch("api/forgot-password", {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        username: username,
        confirmation_code: confirmationCode,
        password: newPassword,
      }),
    })
      .then((response) => response.json())
      .then((data) => {
        if (data.error) {
          setErrorMsg(data.error);
          return;
        }
        if (data.status === "SUCCESS") {
          router.replace("/dashboard");
        }
      })
      .catch((error) => {
        setErrorMsg("Error on Login: " + error);
      });
  };

  return (
    <form
      onSubmit={resetPasswordHandler}
      className="flex flex-col gap-4 items-center border-black border pb-12 px-4 font-body"
    >
      <p className="text-red-600">{errorMsg && errorMsg}</p>
      <label htmlFor="username">username</label>
      <input
        type="text"
        className="bg-main border border-black"
        name="username"
        value={username}
        onChange={(e) => setUsername(e.target.value)}
      />
      <label htmlFor="new-password">new password</label>
      <input
        className="bg-main border border-black"
        name="new-password"
        type="password"
        value={newPassword}
        onChange={(e) => setNewPassword(e.target.value)}
      />
      <label htmlFor="confirmation-code">confirmation code</label>
      <input
        className="bg-main border border-black"
        name="confirmation-code"
        type="text"
        value={confirmationCode}
        onChange={(e) => setConfirmationCode(e.target.value)}
      />
      <button className="px-4 mt-4 border border-black" type="submit">
        Enter
      </button>
    </form>
  );
}
