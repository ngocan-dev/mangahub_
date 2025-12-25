import type { Metadata } from "next";
import { Inter } from "next/font/google";
import type { ReactNode } from "react";
import "./globals.css";

import Navbar from "@/components/Navbar";
import SystemStatusBanner from "@/components/SystemStatusBanner";
import { AuthProvider } from "@/context/AuthContext";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "MangaHub | Read and Discover Manga",
  description: "Modern MangaHub frontend built with Next.js 14 and TypeScript.",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body className={`${inter.className} bg-slate-950 text-slate-100`}>
        <AuthProvider>
          <Navbar />
          <main className="mx-auto max-w-6xl px-4 py-6">
            <SystemStatusBanner />
            {children}
          </main>
        </AuthProvider>
      </body>
    </html>
  );
}