"use client";
import { usePathname } from "next/navigation";
import Link from "next/link";

function generateBreadCrumbs(pathname: string) {
  const paths = pathname.split("/").filter(Boolean);
  const breadCrumbs = [];
  let currentPath = "";

  // TODO: this leads to problems if the user clicks on the breadcrumb
  // TODO: that they are currently on.
  for (let i = 0; i < paths.length; i++) {
    if (!isNaN(+paths[i])) {
      continue;
    }
    currentPath += `/${paths[i]}`;

    // Capitalize the first letter for other path segments
    const label = paths[i].charAt(0).toUpperCase() + paths[i].slice(1);
    breadCrumbs.push({ path: currentPath, label });
  }

  return breadCrumbs;
}

export default function TeacherBreadCrumbs() {
  const pathname = usePathname();
  const breadCrumbs = generateBreadCrumbs(pathname);

  const isActive = (path: string, label: string) => {
    const pathParts = pathname.split("/").filter(Boolean);
    const breadcrumbParts = path.split("/").filter(Boolean);

    // For dashboard
    if (pathname === "/teacher-dashboard" && label === "Dashboard") {
      return true;
    }

    // For course
    if (
      pathParts.length === 3 &&
      pathParts[1] === "course" &&
      label.startsWith("Course")
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
    <nav className="flex flex-row flex-wrap gap-2 items-center">
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
