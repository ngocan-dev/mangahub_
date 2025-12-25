import {
  getChaptersByMangaId,
  getMangaById as fetchMangaById,
  getPopularManga,
  getReviewsByMangaId,
  searchManga as searchMangaApi,
  type Chapter,
  type Manga,
  type Review,
} from "@/service/api";

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
  return getChaptersByMangaId(id);
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
