import { http } from "@/lib/http";

export interface Manga {
  id: number;
  slug: string;
  name: string;
  title: string;
  author: string;
  artist?: string;
  genre: string;
  status: string;
  description: string;
  image: string;
  rating_point: number;
  views?: number;
  relevance_score?: number;
}

export interface ChapterSummary {
  id: number;
  manga_id: number;
  number: number;
  title: string;
  updated_at?: string;
}

export interface LibraryStatus {
  status: string;
  current_chapter?: number;
  started_at?: string;
  completed_at?: string;
}

export interface UserProgress {
  current_chapter: number;
  current_chapter_id?: number;
  progress_percent?: number;
  last_read_at: string;
}

export interface MangaDetail extends Manga {
  chapter_count: number;
  chapters?: ChapterSummary[];
  library_status?: LibraryStatus;
  user_progress?: UserProgress;
}

export interface SearchResponse {
  results: Manga[];
  total: number;
  page: number;
  limit: number;
  pages: number;
}

export interface LibraryEntry {
  manga_id: number;
  title: string;
  cover_image: string;
  status: string;
  started_at?: string;
  completed_at?: string;
  last_updated_at: string;
}

export interface GetLibraryResponse {
  entries: LibraryEntry[];
}

export interface ChapterDetail {
  id: number;
  manga_id: number;
  chapter_number: number;
  title: string;
  content_text: string;
  created_at?: string;
}

export interface Review {
  review_id: number;
  user_id: number;
  username: string;
  manga_id: number;
  rating: number;
  content: string;
  created_at: string;
  updated_at?: string;
}

export interface GetReviewsResponse {
  reviews: Review[];
  total: number;
  page: number;
  limit: number;
  pages: number;
  average_rating: number;
  total_reviews: number;
}

export interface CreateReviewRequest {
  rating: number;
  content: string;
}

export interface CreateReviewResponse {
  message: string;
  review?: Review;
}

export interface AddToLibraryRequest {
  status: string;
  current_chapter?: number;
}

export interface AddToLibraryResponse {
  status: string;
  current_chapter: number;
  already_in_library: boolean;
}

export interface UpdateProgressRequest {
  current_chapter: number;
}

export interface UpdateProgressResponse {
  message: string;
  user_progress?: UserProgress;
  broadcasted: boolean;
}

export async function getPopularManga(limit = 10): Promise<Manga[]> {
  const { data } = await http.get<Manga[]>("/mangas/popular", { params: { limit } });
  return data;
}

export async function searchManga(keyword: string): Promise<SearchResponse> {
  const { data } = await http.get<SearchResponse>("/mangas/search", { params: { q: keyword } });
  return data;
}

export async function getMangaById(id: number): Promise<MangaDetail> {
  const { data } = await http.get<MangaDetail>(`/mangas/${id}`);
  return data;
}

export async function addToLibrary(id: number, payload: AddToLibraryRequest): Promise<AddToLibraryResponse> {
  const { data } = await http.post<AddToLibraryResponse>(`/mangas/${id}/library`, payload);
  return data;
}

export async function updateProgress(id: number, payload: UpdateProgressRequest): Promise<UpdateProgressResponse> {
  const { data } = await http.put<UpdateProgressResponse>(`/mangas/${id}/progress`, payload);
  return data;
}

export async function getReviewsByMangaId(id: number, page = 1, limit = 10): Promise<GetReviewsResponse> {
  const { data } = await http.get<GetReviewsResponse>(`/mangas/${id}/reviews`, { params: { page, limit } });
  return data;
}

export async function createReview(id: number, payload: CreateReviewRequest): Promise<CreateReviewResponse> {
  const { data } = await http.post<CreateReviewResponse>(`/mangas/${id}/reviews`, payload);
  return data;
}

export async function getLibrary(): Promise<GetLibraryResponse> {
  const { data } = await http.get<GetLibraryResponse>("/library");
  return data;
}

export async function getChapterById(id: number): Promise<ChapterDetail> {
  const { data } = await http.get<ChapterDetail>(`/chapters/${id}`);
  return data;
}
