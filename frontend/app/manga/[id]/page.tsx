"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";

import { useAuth } from "@/context/AuthContext";
import {
  mangaService,
  type GetReviewsResponse,
  type MangaDetail,
} from "@/service/manga.service";

interface PageProps {
  params: { id?: string };
}

export default function MangaDetailPage({ params }: PageProps) {
  const mangaId = useMemo(() => Number(params.id), [params.id]);
  const { isAuthenticated } = useAuth();
  const [manga, setManga] = useState<MangaDetail | null>(null);
  const [reviews, setReviews] = useState<GetReviewsResponse | null>(null);
  const [currentChapter, setCurrentChapter] = useState<number>(0);
  const [status, setStatus] = useState<string | null>(null);
  const [inLibrary, setInLibrary] = useState<boolean>(false);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [reviewComment, setReviewComment] = useState<string>("");
  const [reviewRating, setReviewRating] = useState<number>(8);

  const sortedChapters = useMemo(() => {
    if (!manga?.chapters?.length) return [];
    return [...manga.chapters].sort((a, b) => (a.number ?? 0) - (b.number ?? 0));
  }, [manga?.chapters]);

  const totalChapters = useMemo(() => manga?.chapter_count ?? sortedChapters.length ?? 0, [manga?.chapter_count, sortedChapters.length]);

  const progressPercent = useMemo(() => {
    if (!totalChapters || currentChapter <= 0) return 0;
    return Math.min(100, Math.round((currentChapter / totalChapters) * 100));
  }, [currentChapter, totalChapters]);

  useEffect(() => {
    if (!Number.isFinite(mangaId) || mangaId <= 0) {
      setError("Manga not found.");
      return;
    }
    setManga(null);
    setReviews(null);
    setLoading(true);
    const loadData = async () => {
      setError(null);
      try {
        const [mangaDetail, reviewList] = await Promise.all([
          mangaService.getById(mangaId),
          mangaService.getReviews(mangaId),
        ]);
        setManga(mangaDetail);
        setCurrentChapter(mangaDetail.user_progress?.current_chapter ?? mangaDetail.library_status?.current_chapter ?? 0);
        setInLibrary(Boolean(mangaDetail.library_status));
        setReviews(reviewList);
      } catch (err) {
        setError("Unable to load this manga right now. Please ensure the server is online.");
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    void loadData();
  }, [mangaId]);

  const handleAddToLibrary = async () => {
    if (!Number.isFinite(mangaId) || mangaId <= 0) return;
    setStatus(null);
    setError(null);
    try {
      const response = await mangaService.addLibraryEntry(mangaId, { status: "reading", current_chapter: currentChapter });
      setInLibrary(true);
      setStatus(response.already_in_library ? "Already in your library." : "Added to your library!");
      setCurrentChapter(response.current_chapter ?? currentChapter);
    } catch (err) {
      setError("Could not add to library. Please try again.");
      console.error(err);
    }
  };

  const handleProgressUpdate = async () => {
    if (!Number.isFinite(mangaId) || mangaId <= 0) return;
    if (currentChapter < 1 || (totalChapters > 0 && currentChapter > totalChapters)) {
      setError("Select a valid chapter before updating progress.");
      return;
    }
    setStatus(null);
    setError(null);
    try {
      const response = await mangaService.setProgress(mangaId, { current_chapter: currentChapter });
      if (response.user_progress?.current_chapter !== undefined) {
        setCurrentChapter(response.user_progress.current_chapter);
      }
      setStatus(response.message || "Progress updated.");
    } catch (err) {
      setError("Could not update progress.");
      console.error(err);
    }
  };

  const handleSubmitReview = async () => {
    if (!Number.isFinite(mangaId) || mangaId <= 0) return;
    setStatus(null);
    setError(null);
    try {
      await mangaService.submitReview(mangaId, { content: reviewComment, rating: reviewRating });
      setReviewComment("");
      setStatus("Review submitted!");
      const refreshed = await mangaService.getReviews(mangaId);
      setReviews(refreshed);
    } catch (err) {
      setError("Unable to submit review.");
      console.error(err);
    }
  };

  if (loading) {
    return <p className="text-slate-300">Loading manga...</p>;
  }

  if (!manga) {
    return <p className="text-slate-300">{error ?? "Manga not found."}</p>;
  }

  const genres = (manga.genre ?? "")
    .split(",")
    .map((value) => value.trim())
    .filter(Boolean);

  return (
    <section className="space-y-6">
      <div className="rounded-xl border border-slate-800 bg-slate-900 p-6 shadow-lg">
        <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div>
            <h1 className="text-3xl font-semibold text-white">{manga.title}</h1>
            {manga.author ? <p className="text-sm text-slate-400">by {manga.author}</p> : null}
            <p className="mt-3 text-sm text-slate-300">{manga.description ?? "No description available."}</p>
          </div>
          {manga.rating_point ? (
            <div className="rounded-lg bg-slate-800 px-4 py-2 text-center">
              <p className="text-xs text-slate-400">Community rating</p>
              <p className="text-xl font-semibold text-amber-300">⭐ {manga.rating_point}</p>
            </div>
          ) : null}
        </div>
        {genres.length ? (
          <div className="mt-4 flex flex-wrap gap-2">
            {genres.map((genre) => (
              <span key={genre} className="rounded-full bg-slate-800 px-3 py-1 text-xs text-slate-100">
                {genre}
              </span>
            ))}
          </div>
        ) : null}
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <div className="card md:col-span-2">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-white">Reading progress</h2>
            <div className="text-right">
              <p className="text-sm text-slate-400">{progressPercent}%</p>
              <p className="text-xs text-slate-500">
                Chapter {currentChapter || 0}
                {totalChapters ? ` / ${totalChapters}` : ""}
              </p>
            </div>
          </div>
          <input
            type="range"
            min={0}
            max={totalChapters > 0 ? totalChapters : 100}
            value={currentChapter}
            onChange={(e) => setCurrentChapter(Number(e.target.value))}
            className="mt-3 w-full accent-primary"
          />
          <div className="mt-3 flex items-center gap-3">
            <button
              className="btn-primary"
              onClick={handleProgressUpdate}
              disabled={!isAuthenticated || currentChapter < 1 || (totalChapters > 0 && currentChapter > totalChapters)}
            >
              Update progress
            </button>
            {!isAuthenticated ? <p className="text-sm text-amber-300">Login to update your progress.</p> : null}
          </div>
        </div>
        <div className="card space-y-3">
          <h2 className="text-lg font-semibold text-white">Library</h2>
          <p className="text-sm text-slate-300">Add this manga to your personal library.</p>
          <button className="btn-secondary w-full" onClick={handleAddToLibrary} disabled={!isAuthenticated || inLibrary}>
            {inLibrary ? "In library" : "Add to library"}
          </button>
          {!isAuthenticated ? <p className="text-xs text-amber-300">Sign in to save manga to your library.</p> : null}
          {status ? <p className="text-sm text-emerald-300">{status}</p> : null}
          {error ? <p className="text-sm text-red-400">{error}</p> : null}
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <div className="card md:col-span-2">
          <h3 className="text-lg font-semibold text-white">Chapters</h3>
          {sortedChapters.length ? (
            <ul className="mt-3 space-y-2 text-sm text-slate-300">
              {sortedChapters.map((chapter) => (
                <li
                  key={chapter.id}
                  className={`rounded-lg text-slate-200 transition hover:-translate-y-0.5 hover:bg-slate-700 ${
                    currentChapter > 0 && (chapter.number ?? 0) <= currentChapter ? "bg-slate-800/70 text-slate-400" : "bg-slate-800"
                  }`}
                >
                  <Link href={`/chapter/${chapter.id}`} className="flex items-center justify-between px-3 py-2">
                    <span className="font-medium">
                      {chapter.number ? `Ch. ${chapter.number} — ` : ""}
                      {chapter.title}
                    </span>
                    {chapter.updated_at ? (
                      <span className="text-xs text-slate-400">
                        Updated {new Date(chapter.updated_at).toLocaleDateString()}
                      </span>
                    ) : null}
                  </Link>
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-sm text-slate-400">No chapters available.</p>
          )}
        </div>
        <div className="card space-y-3">
          <h3 className="text-lg font-semibold text-white">Reviews</h3>
          <div className="space-y-2 text-sm">
            {reviews?.reviews?.length ? (
              reviews.reviews.map((review) => (
                <div key={review.review_id} className="rounded-lg bg-slate-800 p-3">
                  <div className="flex items-center justify-between text-xs text-slate-400">
                    <span>Rating: {review.rating}/10</span>
                    <span>{review.username}</span>
                  </div>
                  <p className="mt-1 text-slate-200">{review.content}</p>
                </div>
              ))
            ) : (
              <p className="text-slate-400">No reviews yet.</p>
            )}
          </div>
          <div className="space-y-2 rounded-lg border border-slate-800 bg-slate-950 p-3">
            <p className="text-sm font-medium text-white">Add a review</p>
            <label className="text-xs text-slate-400" htmlFor="rating">
              Rating (1-10)
            </label>
            <input
              id="rating"
              type="number"
              min={1}
              max={10}
              value={reviewRating}
              onChange={(e) => setReviewRating(Number(e.target.value))}
            />
            <textarea
              className="w-full"
              rows={3}
              placeholder="Share your thoughts..."
              value={reviewComment}
              onChange={(e) => setReviewComment(e.target.value)}
            />
            <button className="btn-primary w-full" onClick={handleSubmitReview} disabled={!isAuthenticated}>
              Submit review
            </button>
            {!isAuthenticated ? <p className="text-xs text-amber-300">Login to leave a review.</p> : null}
          </div>
        </div>
      </div>
    </section>
  );
}
