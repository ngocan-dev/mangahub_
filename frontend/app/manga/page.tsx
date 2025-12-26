"use client";

import { useEffect, useMemo, useState } from "react";

import MangaCard from "../components/MangaCard";
import { mangaService } from "@/service/manga.service";
import type { Manga } from "@/service/api";

type Mode = "popular" | "search";

function normalizePaged<T>(raw: any, page: number, limit: number, fallbackKey?: string) {
  // Case 1: backend trả thẳng array (hiện tại /popular của bạn đang vậy)
  if (Array.isArray(raw)) {
    return {
      items: raw as T[],
      // nếu trả đủ "limit" thì tạm coi như còn trang nữa (ước lượng)
      hasMore: (raw as T[]).length === limit,
      page,
      limit,
    };
  }

  // Case 2: backend trả dạng paged chuẩn
  if (raw && typeof raw === "object") {
    const items =
      (raw.items as T[]) ??
      (fallbackKey ? (raw[fallbackKey] as T[]) : undefined) ??
      [];

    const hasMore =
      typeof raw.has_more === "boolean"
        ? raw.has_more
        : items.length === limit;

    return {
      items,
      hasMore,
      page: Number(raw.page ?? page),
      limit: Number(raw.limit ?? limit),
    };
  }

  return { items: [] as T[], hasMore: false, page, limit };
}

export default function MangaListPage() {
  const [manga, setManga] = useState<Manga[]>([]);
  const [query, setQuery] = useState<string>("");

  const [mode, setMode] = useState<Mode>("popular");

  const [page, setPage] = useState<number>(1);
  const [hasMore, setHasMore] = useState<boolean>(true);

  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const limit = 20;

  const trimmedQuery = useMemo(() => query.trim(), [query]);

  const loadPopular = async (opts?: { reset?: boolean }) => {
    const reset = opts?.reset ?? false;

    setLoading(true);
    setError(null);

    const nextPage = reset ? 1 : page;

    try {
      const raw = await mangaService.getPopular(nextPage, limit);
      const { items, hasMore: more } = normalizePaged<Manga>(raw, nextPage, limit);

      setManga((prev) => (reset ? items : [...prev, ...items]));
      setHasMore(more);
      setPage(nextPage + 1);
      setMode("popular");
    } catch (err) {
      setError("Unable to load popular manga right now. Please check your connection to the server.");
      if (reset) setManga([]);
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const loadSearch = async (opts?: { reset?: boolean }) => {
    const reset = opts?.reset ?? false;

    if (!trimmedQuery) {
      // nếu query rỗng thì quay về popular
      return loadPopular({ reset: true });
    }

    setLoading(true);
    setError(null);

    const nextPage = reset ? 1 : page;

    try {
      const raw = await mangaService.search(trimmedQuery, nextPage, limit);

      // search hiện tại bạn đang dùng raw.results; nếu backend nâng cấp thì trả raw.items
      const { items, hasMore: more } = normalizePaged<Manga>(raw, nextPage, limit, "results");

      setManga((prev) => (reset ? items : [...prev, ...items]));
      setHasMore(more);
      setPage(nextPage + 1);
      setMode("search");
    } catch (err) {
      setError("Search failed. Please try again.");
      if (reset) setManga([]);
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = async () => {
    // reset pagination khi search mới
    setPage(1);
    setHasMore(true);
    await loadSearch({ reset: true });
  };

  const handleLoadMore = async () => {
    if (loading || !hasMore) return;
    if (mode === "search") return loadSearch({ reset: false });
    return loadPopular({ reset: false });
  };

  useEffect(() => {
    void loadPopular({ reset: true });
    // eslint-disable-next-line react-hooks/exhaustive-deps
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

      {loading && manga.length === 0 ? (
        <p className="text-slate-300">Loading manga...</p>
      ) : (
        <>
          <div className="grid gap-4 md:grid-cols-2">
            {manga.map((item) => (
              <MangaCard key={item.id} manga={item} />
            ))}
          </div>

          {!manga.length && !loading ? (
            <p className="text-slate-400">No manga found. Try adjusting your search.</p>
          ) : null}

          <div className="flex justify-center">
            <button
              onClick={handleLoadMore}
              className="btn-primary"
              disabled={loading || !hasMore}
            >
              {loading ? "Loading..." : hasMore ? "Load more" : "No more results"}
            </button>
          </div>
        </>
      )}
    </section>
  );
}
