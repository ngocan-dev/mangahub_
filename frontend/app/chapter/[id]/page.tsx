"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";

import { useAuth } from "@/context/AuthContext";
import { chapterService, type ChapterDetail } from "@/service/chapter.service";
import { mangaService } from "@/service/manga.service";
import { type ChapterSummary } from "@/service/api";

interface PageProps {
  params: { id?: string };
}

export default function ChapterPage({ params }: PageProps) {
  const chapterId = useMemo(() => Number(params.id), [params.id]);
  const { isAuthenticated } = useAuth();
  const [chapter, setChapter] = useState<ChapterDetail | null>(null);
  const [chapterList, setChapterList] = useState<ChapterSummary[]>([]);
  const [totalChapters, setTotalChapters] = useState<number>(0);
  const [loading, setLoading] = useState(true);
  const [progressUpdating, setProgressUpdating] = useState(false);
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

  useEffect(() => {
    if (!chapter) return;

    const loadMangaMeta = async () => {
      try {
        const mangaDetail = await mangaService.getById(chapter.manga_id);
        setChapterList(mangaDetail.chapters ?? []);
        setTotalChapters(mangaDetail.chapter_count ?? mangaDetail.chapters?.length ?? 0);
      } catch (err) {
        console.error(err);
      }
    };

    void loadMangaMeta();
  }, [chapter?.manga_id, chapter]);

  useEffect(() => {
    if (!chapter || !isAuthenticated) return;

    const syncProgress = async () => {
      try {
        setProgressUpdating(true);
        await mangaService.setProgress(chapter.manga_id, { current_chapter: chapter.chapter_number });
      } catch (err) {
        console.error(err);
      } finally {
        setProgressUpdating(false);
      }
    };

    void syncProgress();
  }, [chapter?.chapter_number, chapter?.id, chapter?.manga_id, isAuthenticated]);

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

  const previousChapter = useMemo(
    () => chapterList.find((item) => item.number === chapter.chapter_number - 1) ?? null,
    [chapter.chapter_number, chapterList],
  );

  const nextChapter = useMemo(
    () => chapterList.find((item) => item.number === chapter.chapter_number + 1) ?? null,
    [chapter.chapter_number, chapterList],
  );

  const showPrevious = chapter.chapter_number > 1 && previousChapter;
  const showNext = totalChapters > 0 && chapter.chapter_number < totalChapters && nextChapter;

  const createdAt = chapter.created_at ? new Date(chapter.created_at).toLocaleString() : null;

  return (
    <section className="space-y-6">
      <div className="flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
        <div className="flex flex-col gap-2">
          <Link href={`/manga/${chapter.manga_id}`} className="text-sm text-primary underline underline-offset-4">
            ← Back to manga
          </Link>
          <div>
            <p className="text-xs uppercase tracking-wide text-slate-400">Chapter {chapter.chapter_number}</p>
            <h1 className="text-3xl font-semibold text-white">{chapter.title}</h1>
            {createdAt ? <p className="text-xs text-slate-500">Published {createdAt}</p> : null}
          </div>
          {progressUpdating ? <p className="text-xs text-slate-500">Updating your progress…</p> : null}
        </div>
        <div className="flex items-center gap-2">
          {showPrevious && previousChapter ? (
            <Link
              href={`/chapter/${previousChapter.id}`}
              className="rounded-lg border border-slate-800 px-3 py-2 text-sm text-slate-200 transition hover:bg-slate-800"
            >
              ← Previous
            </Link>
          ) : null}
          {showNext && nextChapter ? (
            <Link
              href={`/chapter/${nextChapter.id}`}
              className="rounded-lg border border-slate-800 px-3 py-2 text-sm text-slate-200 transition hover:bg-slate-800"
            >
              Next →
            </Link>
          ) : null}
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
