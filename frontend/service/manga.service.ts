import {
  getChaptersByMangaId,
  getMangaById as fetchMangaById,
  getPopularManga,
  getReviewsByMangaId,
  searchManga as searchMangaApi,
} from "@/service/api";
import type { Chapter, Manga, Review } from "@/service/api";

const isChapterArray = (data: unknown): data is Chapter[] => Array.isArray(data);

const normalizeChapter = (chapter: unknown): Chapter | null => {
  if (!chapter || typeof chapter !== "object") return null;
  const raw = chapter as Record<string, unknown>;

  const id = Number(raw.id);
  const mangaId = Number(raw.manga_id);
  const chapterNumber = Number(raw.chapter_number);
  const title = typeof raw.title === "string" ? raw.title : "Untitled";
  const contentUrl = typeof raw.content_url === "string" ? raw.content_url : "";

  if (!Number.isFinite(id) || !Number.isFinite(mangaId)) return null;

  return {
    id,
    manga_id: mangaId,
    chapter_number: Number.isFinite(chapterNumber) ? chapterNumber : 0,
    title,
    content_url: contentUrl,
  };
};

const normalizeChapters = (payload: unknown): Chapter[] => {
  const candidates = (() => {
    if (isChapterArray(payload)) return payload;
    if (payload && typeof payload === "object" && "chapters" in payload) {
      const rawChapters = (payload as { chapters?: unknown }).chapters;
      if (isChapterArray(rawChapters)) return rawChapters;
    }
    return [] as unknown[];
  })();

  return candidates
    .map((chapter) => normalizeChapter(chapter))
    .filter((chapter): chapter is Chapter => Boolean(chapter));
};

async function getPopular(): Promise<Manga[]> {
  return getPopularManga();
}

async function searchManga(query: string): Promise<Manga[]> {
  return searchMangaApi(query);
}

async function getMangaById(id: number): Promise<Manga> {
  return fetchMangaById(id);
}

async function getChapters(id: number): Promise<Chapter[]> {
  const raw = await getChaptersByMangaId(id);
  return normalizeChapters(raw);
}

async function getReviews(id: number): Promise<Review[]> {
  return getReviewsByMangaId(id);
}

export const mangaService = {
  getPopular,
  searchManga,
  getMangaById,
  getChapters,
  getReviews,
};
