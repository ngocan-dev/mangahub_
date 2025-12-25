import { http } from "@/lib/http";

export interface Manga {
  id: string;
  title: string;
  author?: string;
  description?: string;
  coverImage?: string;
  rating?: number;
  genres?: string[];
  status?: string;
}

export interface Chapter {
  id: string;
  title: string;
  number?: number;
  releasedAt?: string;
}

export interface Review {
  id: string;
  userId?: string;
  rating: number;
  comment: string;
  createdAt?: string;
}

export interface ProgressPayload {
  progress: number;
  chapterId?: string;
}

export interface ReviewPayload {
  rating: number;
  comment: string;
}

async function getPopular(): Promise<Manga[]> {
  const { data } = await http.get<Manga[]>("/manga/popular");
  return data;
}

async function searchManga(query: string): Promise<Manga[]> {
  const { data } = await http.get<Manga[]>("/manga/search", { params: { query } });
  return data;
}

async function getMangaById(id: string): Promise<Manga> {
  const { data } = await http.get<Manga>(`/manga/${id}`);
  return data;
}

async function getChapters(id: string): Promise<Chapter[]> {
  const { data } = await http.get<Chapter[]>(`/chapters/${id}`);
  return data;
}

async function addToLibrary(id: string): Promise<void> {
  await http.post(`/manga/${id}/library`);
}

async function updateProgress(id: string, payload: ProgressPayload): Promise<void> {
  await http.put(`/manga/${id}/progress`, payload);
}

async function addReview(id: string, payload: ReviewPayload): Promise<void> {
  await http.post(`/manga/${id}/reviews`, payload);
}

async function getReviews(id: string): Promise<Review[]> {
  const { data } = await http.get<Review[]>(`/manga/${id}/reviews`);
  return data;
}

export const mangaService = {
  getPopular,
  searchManga,
  getMangaById,
  getChapters,
  addToLibrary,
  updateProgress,
  addReview,
  getReviews,
};