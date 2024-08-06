"use client";
import { usePathname } from "next/navigation";
import Link from "next/link";

function generateBreadCrumbs(pathname: string) {
  const paths = pathname.split("/").filter(Boolean);
  const breadCrumbs = [];
  let currentPath = "";

  for (let i = 0; i < paths.length; i++) {
    currentPath += `/${paths[i]}`;

    if (paths[i] === "book") {
      // Combine "book" with the next number
      if (i + 1 < paths.length && !isNaN(Number(paths[i + 1]))) {
        const label = `Book ${paths[i + 1]}`;
        currentPath += `/${paths[i + 1]}`;
        breadCrumbs.push({ path: currentPath, label });
        i++; // Skip the next iteration (the number)
      }
    } else if (paths[i] === "chapter") {
      // Combine "chapter" with the next number, include "activities" in path but not label
      if (
        i + 2 < paths.length &&
        !isNaN(Number(paths[i + 1])) &&
        paths[i + 2] === "activities"
      ) {
        const label = `Chapter ${paths[i + 1]}`;
        currentPath += `/${paths[i + 1]}/activities`;
        breadCrumbs.push({ path: currentPath, label });
        i += 2; // Skip the next two iterations (the number and "activities")
      }
    } else if (paths[i] !== "activities") {
      // Capitalize the first letter for other path segments
      const label = paths[i].charAt(0).toUpperCase() + paths[i].slice(1);
      breadCrumbs.push({ path: currentPath, label });
    }
  }

  return breadCrumbs;
}

export default function BreadCrumbs() {
  const pathname = usePathname();
  const breadCrumbs = generateBreadCrumbs(pathname);

  const isActive = (path: string, label: string) => {
    const pathParts = pathname.split("/").filter(Boolean);
    const breadcrumbParts = path.split("/").filter(Boolean);

    // For dashboard
    if (pathname === "/dashboard" && label === "Dashboard") {
      return true;
    }

    // For book
    if (
      pathParts.length === 3 &&
      pathParts[1] === "book" &&
      label.startsWith("Book")
    ) {
      return true;
    }

    // For chapter
    if (
      pathParts.length === 5 &&
      pathParts[3] === "chapter" &&
      label.startsWith("Chapter")
    ) {
      return true;
    }

    // For activities or other pages
    if (
      pathParts.length > 5 &&
      breadcrumbParts.length === pathParts.length &&
      breadcrumbParts[breadcrumbParts.length - 1] ===
        pathParts[pathParts.length - 1]
    ) {
      return true;
    }

    return false;
  };

  return (
    <nav className="flex flex-row gap-2 items-center">
      {breadCrumbs.map((breadCrumb, index) => (
        <div key={index} className="flex items-center">
          <Link
            href={breadCrumb.path}
            className={`hover:underline ${isActive(breadCrumb.path, breadCrumb.label) ? "underline font-semibold" : ""}`}
          >
            {breadCrumb.label}
          </Link>
          {index < breadCrumbs.length - 1 && (
            <span className="mx-2 text-gray-500">&gt;</span>
          )}
        </div>
      ))}
    </nav>
  );
}
