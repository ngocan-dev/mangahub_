"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";

import { useAuth } from "@/context/AuthContext";

const navLinks = [
  { href: "/manga", label: "Manga" },
  { href: "/library", label: "Library", protected: true },
  { href: "/friends", label: "Friends", protected: true },
  { href: "/analytics", label: "Analytics", protected: true },
];

export default function Navbar() {
  const pathname = usePathname();
  const router = useRouter();
  const { isAuthenticated, logout, user } = useAuth();

  const handleLogout = () => {
    logout();
    router.push("/login");
  };

  return (
    <nav className="sticky top-0 z-20 w-full border-b border-slate-800 bg-slate-950/90 backdrop-blur">
      <div className="mx-auto flex max-w-6xl items-center justify-between px-4 py-4">
        <Link href="/manga" className="text-lg font-semibold text-white">
          MangaHub
        </Link>
        <div className="flex items-center gap-4 text-sm font-medium">
          {navLinks.map(({ href, label, protected: isProtected }) => {
            if (isProtected && !isAuthenticated) return null;
            const active = pathname.startsWith(href);
            return (
              <Link
                key={href}
                href={href}
                className={`rounded-md px-3 py-2 transition-colors ${
                  active ? "bg-slate-800 text-white" : "text-slate-300 hover:bg-slate-900 hover:text-white"
                }`}
              >
                {label}
              </Link>
            );
          })}
          {!isAuthenticated ? (
            <>
              <Link href="/login" className="rounded-md px-3 py-2 text-slate-300 hover:bg-slate-900 hover:text-white">
                Login
              </Link>
              <Link href="/register" className="btn-primary">
                Register
              </Link>
            </>
          ) : (
            <div className="flex items-center gap-3">
              <span className="hidden rounded-full bg-slate-800 px-3 py-1 text-xs text-slate-200 sm:block">
                {user?.username ?? user?.email ?? "Signed in"}
              </span>
              <button onClick={handleLogout} className="btn-secondary">
                Logout
              </button>
            </div>
          )}
        </div>
      </div>
    </nav>
  );
}
