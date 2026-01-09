"use client";

import Link from "next/link";
import { useSession, signIn, signOut } from "next-auth/react";

export function Navbar() {
  const { data: session, status } = useSession();

  return (
    <nav className="flex items-center justify-between py-4 border-b border-slate-700">
      <Link href="/" className="text-2xl font-bold text-white">
        T3 Blog
      </Link>
      <div className="flex items-center gap-4">
        {status === "loading" ? (
          <span className="text-gray-400">Loading...</span>
        ) : session ? (
          <>
            <Link href="/dashboard" className="text-gray-300 hover:text-white">
              Dashboard
            </Link>
            <Link href="/new" className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700">
              New Post
            </Link>
            <button
              onClick={() => signOut()}
              className="text-gray-300 hover:text-white"
            >
              Sign Out
            </button>
          </>
        ) : (
          <button
            onClick={() => signIn("google")}
            className="px-4 py-2 bg-white text-slate-900 rounded-md hover:bg-gray-100"
          >
            Sign In
          </button>
        )}
      </div>
    </nav>
  );
}
