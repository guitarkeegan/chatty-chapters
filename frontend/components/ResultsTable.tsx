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

export default function ResultsTable({ results }: { results: Result[] }) {
  return (
    <Table>
      <TableCaption>Chapter Results</TableCaption>
      <TableHeader>
        <TableRow>
          <TableHead className="w-[100px]">Student</TableHead>
          <TableHead>ID</TableHead>
          <TableHead>Chat Score</TableHead>
          <TableHead>Quiz Score</TableHead>
          <TableHead className="text-right">Total Activity Score</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {results.map((result) => (
          <TableRow key={result.user_id}>
            <TableCell className="font-medium">
              {result.first_name} {result.last_name}
            </TableCell>
            <TableCell>{result.user_id}</TableCell>
            <TableCell>{result.chat_score}</TableCell>
            <TableCell>{result.quiz_score}</TableCell>
            <TableCell className="text-right">
              {result.total_activity_score}
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
