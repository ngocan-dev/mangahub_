"use client";

import { useEffect, useMemo, useState } from "react";

import { useAuth } from "@/context/AuthContext";
import { mangaService, type Chapter, type Manga, type Review } from "@/services/manga.service";

interface PageProps {
  params: { id?: string };
}

const persistLibraryId = (id: string) => {
  if (typeof window === "undefined") return;
  const current = JSON.parse(localStorage.getItem("libraryIds") ?? "[]") as string[];
  if (!current.includes(id)) {
    localStorage.setItem("libraryIds", JSON.stringify([...current, id]));
  }
};

export default function MangaDetailPage({ params }: PageProps) {
  const id = params.id;
  const { isAuthenticated } = useAuth();
  const [manga, setManga] = useState<Manga | null>(null);
  const [chapters, setChapters] = useState<Chapter[]>([]);
  const [reviews, setReviews] = useState<Review[]>([]);
  const [progress, setProgress] = useState<number>(0);
  const [status, setStatus] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [reviewComment, setReviewComment] = useState<string>("");
  const [reviewRating, setReviewRating] = useState<number>(8);

  const sortedChapters = useMemo(
    () => [...chapters].sort((a, b) => (a.number ?? 0) - (b.number ?? 0)),
    [chapters],
  );

  useEffect(() => {
    if (!id) {
      setError("Manga not found.");
      return;
    }
    setManga(null);
    setChapters([]);
    setReviews([]);
    const loadData = async () => {
      setError(null);
      try {
        const [mangaDetail, chapterList, reviewList] = await Promise.all([
          mangaService.getMangaById(id),
          mangaService.getChapters(id),
          mangaService.getReviews(id),
        ]);
        setManga(mangaDetail);
        setChapters(chapterList);
        setReviews(reviewList);
      } catch (err) {
        setError("Unable to load this manga right now.");
        console.error(err);
      }
    };
    void loadData();
  }, [id]);

  const handleAddToLibrary = async () => {
    if (!id) return;
    setStatus(null);
    try {
      await mangaService.addToLibrary(id);
      persistLibraryId(id);
      setStatus("Added to your library!");
    } catch (err) {
      setError("Could not add to library. Please try again.");
      console.error(err);
    }
  };

  const handleProgressUpdate = async () => {
    if (!id) return;
    setStatus(null);
    try {
      await mangaService.updateProgress(id, { progress });
      setStatus("Progress updated.");
    } catch (err) {
      setError("Could not update progress.");
      console.error(err);
    }
  };

  const handleSubmitReview = async () => {
    if (!id) return;
    setStatus(null);
    try {
      await mangaService.addReview(id, { rating: reviewRating, comment: reviewComment });
      setReviewComment("");
      setStatus("Review submitted!");
      const refreshed = await mangaService.getReviews(id);
      setReviews(refreshed);
    } catch (err) {
      setError("Unable to submit review.");
      console.error(err);
    }
  };

  if (!manga) {
    return <p className="text-slate-300">{error ?? "Loading manga..."}</p>;
  }

  return (
    <section className="space-y-6">
      <div className="rounded-xl border border-slate-800 bg-slate-900 p-6 shadow-lg">
        <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div>
            <h1 className="text-3xl font-semibold text-white">{manga.title}</h1>
            {manga.author ? <p className="text-sm text-slate-400">by {manga.author}</p> : null}
            <p className="mt-3 text-sm text-slate-300">{manga.description ?? "No description available."}</p>
          </div>
          {manga.rating ? (
            <div className="rounded-lg bg-slate-800 px-4 py-2 text-center">
              <p className="text-xs text-slate-400">Community rating</p>
              <p className="text-xl font-semibold text-amber-300">⭐ {manga.rating}</p>
            </div>
          ) : null}
        </div>
        {manga.genres?.length ? (
          <div className="mt-4 flex flex-wrap gap-2">
            {manga.genres.map((genre) => (
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
            <span className="text-sm text-slate-400">{progress}%</span>
          </div>
          <input
            type="range"
            min={0}
            max={100}
            value={progress}
            onChange={(e) => setProgress(Number(e.target.value))}
            className="mt-3 w-full accent-primary"
          />
          <div className="mt-3 flex items-center gap-3">
            <button className="btn-primary" onClick={handleProgressUpdate} disabled={!isAuthenticated}>
              Update progress
            </button>
            {!isAuthenticated ? <p className="text-sm text-amber-300">Login to update your progress.</p> : null}
          </div>
        </div>
        <div className="card space-y-3">
          <h2 className="text-lg font-semibold text-white">Library</h2>
          <p className="text-sm text-slate-300">Add this manga to your personal library.</p>
          <button className="btn-secondary w-full" onClick={handleAddToLibrary} disabled={!isAuthenticated}>
            Add to library
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
                <li key={chapter.id} className="flex items-center justify-between rounded-lg bg-slate-800 px-3 py-2">
                  <span>
                    {chapter.number ? `Ch. ${chapter.number} — ` : ""}
                    {chapter.title}
                  </span>
                  {chapter.releasedAt ? <span className="text-xs text-slate-500">{chapter.releasedAt}</span> : null}
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
            {reviews.length ? (
              reviews.map((review) => (
                <div key={review.id} className="rounded-lg bg-slate-800 p-3">
                  <div className="flex items-center justify-between text-xs text-slate-400">
                    <span>Rating: {review.rating}/10</span>
                    {review.createdAt ? <span>{new Date(review.createdAt).toLocaleDateString()}</span> : null}
                  </div>
                  <p className="mt-1 text-slate-200">{review.comment}</p>
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
