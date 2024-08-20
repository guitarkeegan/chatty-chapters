"use client";
import { useRouter } from "next/navigation";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import BreadCrumbs from "@/components/BreadCrumbs";

export default function Navbar() {
  const router = useRouter();

  function logout() {
    fetch("/api/user/logout", {
      method: "GET",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      credentials: "include",
    })
      .then((res) => res.json())
      .then((data) => {
        if (data.logout) {
          router.replace("/");
        }
      })
      .catch((error) => console.error("unable to logout: ", error));
  }

  return (
    <header className="bg-main py-12 px-12 flex flex-row justify-between items-center">
      {/* <button type="button" onClick={() => router.push("/")}>
        Home
      </button> */}
      <BreadCrumbs />
      <DropdownMenu>
        <DropdownMenuTrigger>Menu</DropdownMenuTrigger>
        <DropdownMenuContent>
          <DropdownMenuLabel>My Account</DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem>
            <button onClick={logout}>Logout</button>
          </DropdownMenuItem>
          <DropdownMenuItem>
            <button onClick={() => router.push("/dashboard/my-progress")}>
              Course Progress
            </button>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </header>
  );
}
