import type { Metadata } from "next";
import { Antic_Didone, Chivo } from "next/font/google";
import { cn } from "@/lib/utils";
import "./globals.css";

const antic = Antic_Didone({
  weight: "400",
  style: "normal",
  display: "auto",
  preload: true,
  fallback: ["Arial", "sans-serif"],
  adjustFontFallback: true,
  subsets: ["latin"],
});

const chivo = Chivo({
  subsets: ["latin"],
  fallback: ["Helvetica Neue", "Arial", "sans-serif"],
  weight: "400",
  style: "normal",
});

export const metadata: Metadata = {
  title: "Chatty Chapters",
  description:
    "An app for students to chat with their textbook, take quizzes and more!",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={cn("min-h-screen bg-main antialiased", antic, chivo)}>
        {children}
      </body>
    </html>
  );
}
