"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import CourseCard from "@/components/CourseCard";
import { type CourseCardProps } from "@/components/CourseCard";
export default function Progress() {
  const [courses, setCourses] = useState<CourseCardProps[]>();

  const router = useRouter();
  // get the progress for each of the user's courses
  useEffect(() => {
    fetch("/api/user/progress/all", {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
    })
      .then((res) => res.json())
      .then((data) => {
        if (data.error === "invalid or missing authentication token") {
          router.replace("/");
          return;
        }
        setCourses(data.user_progress);
        console.log(data);
      })
      .catch((err) => {
        console.error(err);
      });
  }, []);

  return (
    <section className="flex flex-col justify-center items-center gap-12 py-12 ">
      <h1 className="font-display text-4xl">My Courses</h1>

      {courses ? (
        courses.map((course, idx) => (
          <CourseCard
            key={idx}
            course_id={course.course_id}
            course_title={course.course_title}
            semester={course.semester}
            year={course.year}
            book_title={course.book_title}
            book_id={course.book_id}
          />
        ))
      ) : (
        <div> no courses to view </div>
      )}
    </section>
  );
}
