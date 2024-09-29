"use client";
import CourseCard from "@/components/CourseCard";
import { CourseCardProps } from "@/components/CourseCard";
import { useEffect, useState } from "react";
import { usePathname, useRouter } from "next/navigation";

const TeacherDashboardPage: React.FC = () => {
  const [dashboardData, setDashboardData] = useState<CourseCardProps[]>();
  const [errorMsg, setErrorMsg] = useState<string>();
  const router = useRouter();

  const path = usePathname().split("/");

  useEffect(() => {
    fetch("/api/teacher/dashboard", {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
    })
      .then((res) => {
        console.log(res.status);
        if (res.status === 401) {
          router.replace("/dashboard");
        }
        return res.json();
      })
      .then((data) => {
        setDashboardData(data.courses);
      })
      .catch((error) => setErrorMsg(error));
  }, []);

  return (
    <section className="flex flex-col justify-center items-center gap-12 py-12 ">
      <h1 className="font-display text-4xl">Teacher Dashboard</h1>
      <p className="text-red-600">{errorMsg && errorMsg}</p>

      {dashboardData ? (
        dashboardData.map((course, idx) => (
          <CourseCard
            key={idx}
            path={path}
            course_id={course.course_id}
            course_title={course.course_title}
            semester={course.semester}
            year={course.year}
            book_title={course.book_title}
            book_id={course.book_id}
          />
        ))
      ) : (
        <div>loading courses...</div>
      )}
    </section>
  );
};

export default TeacherDashboardPage;
