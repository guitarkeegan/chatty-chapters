"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";

export default function LoginForm() {
  const [email, setEmail] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [newPassword, setNewPassword] = useState<string>("");
  const [newPasswordVerify, setNewPasswordVerify] = useState<string>("");
  const [errorMsg, setErrorMsg] = useState<string>("");
  const [challenge, setChallenge] = useState<boolean>(false);
  const [signup, setSignup] = useState<boolean>(false);

  const router = useRouter();

  const newPasswordHandler = (e: React.FormEvent) => {
    e.preventDefault();
    if (newPassword !== newPasswordVerify) {
      setErrorMsg("Passwords do not match");
      return;
    }
    fetch("api/new-password", {
      method: "POST",
      credentials: "include",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        username: email,
        password: newPassword,
        confirm_password: newPasswordVerify,
      }),
    })
      .then((response) => response.json())
      .then((data) => {
        if (data.error) {
          setErrorMsg(data.error);
          return;
        }
        router.replace("teacher-dashboard");
      })
      .catch((error) => {
        setErrorMsg("Error on Login: " + error);
      });
  };

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
        if (data.challenge.status === "NEW_PASSWORD_REQUIRED") {
          setChallenge(true);
          setPassword("");
          return;
        }
        router.replace("teacher-dashboard");
      })
      .catch((error) => {
        setErrorMsg("Error on Login: " + error);
      });
  };

  return (
    <>
      {challenge ? (
        <form
          onSubmit={newPasswordHandler}
          className="flex flex-col gap-4 items-center border-black border pt-20 pb-12 px-4 font-body"
        >
          <p className="text-red-600">{errorMsg && errorMsg}</p>
          <label htmlFor="new-password">new password</label>
          <input
            className="bg-main border border-black"
            name="new-password"
            type="password"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
          />
          <label htmlFor="new-password-verify">new password again</label>
          <input
            className="bg-main border border-black"
            name="new-password-verify"
            type="password"
            value={newPasswordVerify}
            onChange={(e) => setNewPasswordVerify(e.target.value)}
          />
          <button className="px-4 mt-4 border border-black" type="submit">
            Enter
          </button>
        </form>
      ) : (
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
          {!signup ? (
            <button
              className="px-4 mt-4 border border-black"
              type="button"
              onClick={() => setSignup(!signup)}
            >
              Request Access
            </button>
          ) : (
            <p className="font-display">
              Send a request email to keegananglim@gmail.com
            </p>
          )}
        </form>
      )}
    </>
  );
}
