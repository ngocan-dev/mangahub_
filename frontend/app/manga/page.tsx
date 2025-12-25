"use client";

import { useEffect, useState } from "react";

import MangaCard from "../components/MangaCard";
import { mangaService } from "@/service/manga.service";
import type { Manga } from "@/service/api";

export default function MangaListPage() {
  const [manga, setManga] = useState<Manga[]>([]);
  const [query, setQuery] = useState<string>("");
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const fetchPopular = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await mangaService.getPopular();
      setManga(data);
    } catch (err) {
      setError("Unable to load popular manga right now. Please check your connection to the server.");
      setManga([]);
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = async () => {
    if (!query.trim()) return fetchPopular();
    setLoading(true);
    setError(null);
    try {
      const data = await mangaService.search(query);
      setManga(data.results ?? []);
    } catch (err) {
      setError("Search failed. Please try again.");
      setManga([]);
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void fetchPopular();
  }, []);

  return (
    <section className="space-y-6">
      <header className="flex flex-col gap-3 rounded-xl border border-slate-800 bg-slate-900 p-4 shadow-lg md:flex-row md:items-center md:justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-white">Discover Manga</h1>
          <p className="text-sm text-slate-400">Browse what&apos;s popular right now or search by title.</p>
        </div>
        <div className="flex w-full flex-col gap-2 md:w-96 md:flex-row">
          <input
            className="w-full"
            placeholder="Search manga..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleSearch()}
          />
          <button onClick={handleSearch} className="btn-primary w-full md:w-auto" disabled={loading}>
            Search
          </button>
        </div>
      </header>

      {error ? <p className="text-red-400">{error}</p> : null}
      {loading ? (
        <p className="text-slate-300">Loading manga...</p>
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          {manga.map((item) => (
            <MangaCard key={item.id} manga={item} />
          ))}
          {!manga.length && <p className="text-slate-400">No manga found. Try adjusting your search.</p>}
        </div>
      )}
    </section>
  );
}
