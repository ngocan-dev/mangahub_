import {
  addToLibrary,
  createReview,
  getLibrary,
  getMangaById,
  getPopularManga,
  getReviewsByMangaId,
  searchManga,
  updateProgress,
  type AddToLibraryRequest,
  type AddToLibraryResponse,
  type CreateReviewRequest,
  type GetLibraryResponse,
  type GetReviewsResponse,
  type Manga,
  type MangaDetail,
  type SearchResponse,
  type UpdateProgressRequest,
  type UpdateProgressResponse,
} from "@/service/api";

async function getPopular(): Promise<Manga[]> {
  return getPopularManga();
}

async function search(query: string): Promise<SearchResponse> {
  return searchManga(query);
}

async function getById(id: number): Promise<MangaDetail> {
  return getMangaById(id);
}

async function addLibraryEntry(id: number, payload: AddToLibraryRequest): Promise<AddToLibraryResponse> {
  return addToLibrary(id, payload);
}

async function setProgress(id: number, payload: UpdateProgressRequest): Promise<UpdateProgressResponse> {
  return updateProgress(id, payload);
}

async function getReviews(id: number, page = 1, limit = 10): Promise<GetReviewsResponse> {
  return getReviewsByMangaId(id, page, limit);
}

async function submitReview(id: number, payload: CreateReviewRequest) {
  return createReview(id, payload);
}

async function getLibraryEntries(): Promise<GetLibraryResponse> {
  return getLibrary();
}

export const mangaService = {
  getPopular,
  search,
  getById,
  addLibraryEntry,
  setProgress,
  getReviews,
  submitReview,
  getLibraryEntries,
};

export type {
  Manga,
  MangaDetail,
  SearchResponse,
  AddToLibraryRequest,
  AddToLibraryResponse,
  UpdateProgressRequest,
  UpdateProgressResponse,
  GetReviewsResponse,
  CreateReviewRequest,
  GetLibraryResponse,
  Review,
} from "@/service/api";
