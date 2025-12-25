"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";

import { chapterService, type ChapterDetail } from "@/service/chapter.service";

interface PageProps {
  params: { id?: string };
}

export default function ChapterPage({ params }: PageProps) {
  const chapterId = useMemo(() => Number(params.id), [params.id]);
  const [chapter, setChapter] = useState<ChapterDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!Number.isFinite(chapterId) || chapterId <= 0) {
      setError("Invalid chapter.");
      setLoading(false);
      return;
    }

    const loadChapter = async () => {
      setLoading(true);
      setError(null);
      try {
        const data = await chapterService.getById(chapterId);
        setChapter(data);
      } catch (err) {
        console.error(err);
        setError("Chapter not found.");
      } finally {
        setLoading(false);
      }
    };

    void loadChapter();
  }, [chapterId]);

  if (loading) {
    return <p className="text-slate-300">Loading chapter...</p>;
  }

  if (error) {
    return (
      <div className="space-y-3 text-slate-200">
        <p className="text-lg font-semibold text-red-400">{error}</p>
        <Link href="/" className="text-primary underline underline-offset-4">
          Go back home
        </Link>
      </div>
    );
  }

  if (!chapter) {
    return <p className="text-slate-300">Chapter not found.</p>;
  }

  const createdAt = chapter.created_at ? new Date(chapter.created_at).toLocaleString() : null;

  return (
    <section className="space-y-6">
      <div className="flex flex-col gap-2">
        <Link href={`/manga/${chapter.manga_id}`} className="text-sm text-primary underline underline-offset-4">
          ← Back to manga
        </Link>
        <div>
          <p className="text-xs uppercase tracking-wide text-slate-400">Chapter {chapter.chapter_number}</p>
          <h1 className="text-3xl font-semibold text-white">{chapter.title}</h1>
          {createdAt ? <p className="text-xs text-slate-500">Published {createdAt}</p> : null}
        </div>
      </div>

      <article className="mx-auto max-w-3xl rounded-xl border border-slate-800 bg-slate-900 p-6 shadow-lg">
        <div className="whitespace-pre-wrap text-base leading-relaxed text-slate-100">
          {chapter.content_text || "This chapter does not have any content yet."}
        </div>
      </article>
    </section>
  );
}
