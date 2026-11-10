# Chatty Chapters Frontend

The Next.js frontend for Chatty Chapters. See the [root README](../README.md) for what
the app does and how the pieces fit together.

## Stack

Next.js 14 (App Router), React 18, TypeScript, Tailwind CSS, and
[shadcn/ui](https://ui.shadcn.com) components on Radix primitives.

## How it talks to the API

- REST calls go to `/api/*`, which `next.config.mjs` rewrites to `NEXT_PUBLIC_BASE_URL`.
  Proxying keeps the API's auth cookies same-origin.
- Chapter chat opens a WebSocket to `NEXT_PUBLIC_WS_BASE_URL` and renders the reply as it
  streams in.
- When the API reports a missing or expired token, pages redirect to the login screen.
  Students who try to open the teacher dashboard are redirected too.

## Pages

```
/                              student login and sign-up
/teacher-login                 teacher login and sign-up
/forgot                        password reset
/dashboard                     the student's books
/dashboard/book/[bookid]       chapters in a book
  .../chapter/[chapterid]/activities        chat, quiz, and listening (placeholder)
/dashboard/my-progress         the student's chat and quiz results
/teacher-dashboard             the teacher's courses
  /course/[course_id]/chapter/[chapter_id]/results   per-student results, chats, and quiz feedback
```

## Running locally

Start the API first (see [`../backend`](../backend)), then:

```sh
cp .env.example .env.local
pnpm install
pnpm dev        # http://localhost:3000
```
