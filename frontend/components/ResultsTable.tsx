"use client";
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { Result } from "@/app/teacher-dashboard/course/[course_id]/chapter/[chapter_id]/results/page";

export default function ResultsTable({
  results,
  rowClick,
}: {
  results: Result[];
  rowClick: (result: Result) => void;
}) {
  return (
    <Table className="border border-gray-500 text-gray-600">
      <TableCaption>Chapter Results</TableCaption>
      <TableHeader>
        <TableRow className="border border-gray-500">
          <TableHead className="w-[100px]">Student</TableHead>
          <TableHead>ID</TableHead>
          <TableHead>Chat Score</TableHead>
          <TableHead>Quiz Score</TableHead>
          <TableHead>Total Activity Score</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {results.map((result) => (
          <TableRow
            key={result.user_id}
            className="cursor-pointer border border-gray-500"
            onClick={() => rowClick(result)}
          >
            <TableCell className="font-medium">
              {result.first_name} {result.last_name}
            </TableCell>
            <TableCell>{result.user_id}</TableCell>
            <TableCell>{result.chat_score}</TableCell>
            <TableCell>{result.quiz_score}</TableCell>
            <TableCell>{result.total_activity_score}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
