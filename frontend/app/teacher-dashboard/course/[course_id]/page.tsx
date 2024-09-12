"use client";
import { useEffect } from "react";
import { useRouter } from "next/navigation";

type TeacherCourseParams = {
  course_id: number;
};

export default function TeacherCourse({ course_id }: TeacherCourseParams) {
  const router = useRouter();

  useEffect(() => {
    fetch(`/api/teacher/course/${course_id}`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
    })
      .then((res) => res.json())
      .then((data) => {
        console.log(data);
      })
      .catch((error) => console.error(error));
  });

  return <div>teacher course page</div>;
}
