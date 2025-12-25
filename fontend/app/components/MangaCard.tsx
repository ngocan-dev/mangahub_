"use client";

import Link from "next/link";

import type { Manga } from "@/services/manga.service";

interface MangaCardProps {
  manga: Manga;
  onAddToLibrary?: () => void;
  actionLabel?: string;
}

export default function MangaCard({ manga, onAddToLibrary, actionLabel }: MangaCardProps) {
  return (
    <div className="card flex flex-col gap-3">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h3 className="text-lg font-semibold text-white">{manga.title}</h3>
          {manga.author ? <p className="text-sm text-slate-400">by {manga.author}</p> : null}
        </div>
        {manga.rating ? (
          <span className="rounded-full bg-slate-800 px-3 py-1 text-xs text-amber-300">⭐ {manga.rating}</span>
        ) : null}
      </div>
      {manga.description ? (
        <p className="text-sm text-slate-300 max-h-24 overflow-hidden text-ellipsis">{manga.description}</p>
      ) : null}
      {manga.genres?.length ? (
        <div className="flex flex-wrap gap-2">
          {manga.genres.map((genre) => (
            <span key={genre} className="rounded-full bg-slate-800 px-2 py-1 text-xs text-slate-200">
              {genre}
            </span>
          ))}
        </div>
      ) : null}
      <div className="flex items-center justify-between">
        <Link href={`/manga/${manga.id}`} className="text-sm font-medium text-accent hover:text-secondary">
          View details →
        </Link>
        {onAddToLibrary ? (
          <button onClick={onAddToLibrary} className="btn-primary text-sm">
            {actionLabel ?? "Add to library"}
          </button>
        ) : null}
      </div>
    </div>
  );
}