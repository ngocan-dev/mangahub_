"use client";

import { useEffect, useState } from "react";

import MangaCard from "@/components/MangaCard";
import ProtectedRoute from "@/components/ProtectedRoute";
import { mangaService, type Manga } from "@/services/manga.service";

const getStoredLibraryIds = (): string[] => {
  if (typeof window === "undefined") return [];
  try {
    return JSON.parse(localStorage.getItem("libraryIds") ?? "[]") as string[];
  } catch {
    return [];
  }
};

export default function LibraryPage() {
  const [library, setLibrary] = useState<Manga[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadLibrary = async () => {
      setLoading(true);
      setError(null);
      try {
        const ids = getStoredLibraryIds();
        const results = await Promise.all(ids.map((id) => mangaService.getMangaById(id)));
        setLibrary(results);
      } catch (err) {
        setError("Unable to load your library right now.");
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    void loadLibrary();
  }, []);

  return (
    <ProtectedRoute>
      <section className="space-y-4">
        <div className="rounded-xl border border-slate-800 bg-slate-900 p-4 shadow-lg">
          <h1 className="text-2xl font-semibold text-white">Your Library</h1>
          <p className="text-sm text-slate-400">
            Manga added from detail pages are stored locally and synced to the backend through the library endpoint.
          </p>
        </div>
        {loading ? <p className="text-slate-300">Loading library...</p> : null}
        {error ? <p className="text-red-400">{error}</p> : null}
        {!loading && !library.length ? (
          <p className="text-slate-400">
            Your library is empty. Browse manga and use &quot;Add to library&quot; on a title to track it here.
          </p>
        ) : null}
        <div className="grid gap-4 md:grid-cols-2">
          {library.map((item) => (
            <MangaCard key={item.id} manga={item} />
          ))}
        </div>
      </section>
    </ProtectedRoute>
  );
}