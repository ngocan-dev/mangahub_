"use client";

import { useEffect, useState } from "react";

import ProtectedRoute from "../components/ProtectedRoute"; // Adjusted path
import { mangaService, type GetLibraryResponse } from "@/service/manga.service";

export default function LibraryPage() {
  const [library, setLibrary] = useState<GetLibraryResponse | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadLibrary = async () => {
      setLoading(true);
      setError(null);
      try {
        const entries = await mangaService.getLibraryEntries();
        setLibrary(entries);
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
            Manga added from detail pages are synced to the backend and listed here.
          </p>
        </div>
        {loading ? <p className="text-slate-300">Loading library...</p> : null}
        {error ? <p className="text-red-400">{error}</p> : null}
        {!loading && !library?.entries?.length ? (
          <p className="text-slate-400">
            Your library is empty. Browse manga and use &quot;Add to library&quot; on a title to track it here.
          </p>
        ) : null}
        <div className="grid gap-4 md:grid-cols-2">
          {library?.entries?.map((item) => (
            <div key={item.manga_id} className="card flex flex-col gap-2">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-lg font-semibold text-white">{item.title}</p>
                  <p className="text-xs text-slate-400 capitalize">{item.status}</p>
                </div>
                <span className="text-xs text-slate-500">
                  Updated {new Date(item.last_updated_at).toLocaleDateString()}
                </span>
              </div>
              <div className="text-sm text-slate-300">
                {item.started_at ? <p>Started: {new Date(item.started_at).toLocaleDateString()}</p> : null}
                {item.completed_at ? <p>Completed: {new Date(item.completed_at).toLocaleDateString()}</p> : null}
              </div>
            </div>
          ))}
        </div>
      </section>
    </ProtectedRoute>
  );
}
