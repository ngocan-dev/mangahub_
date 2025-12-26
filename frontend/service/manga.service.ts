import { http } from "@/lib/http";

import {
  addToLibrary,
  createReview,
  getLibrary,
  getMangaById,
  getPopularManga, // giữ lại để fallback nếu cần
  getReviewsByMangaId,
  searchManga,     // endpoint search đang đúng nằm ở đây
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

export type PagedResponse<T> = {
  items: T[];
  page: number;
  limit: number;
  has_more: boolean;
};

function normalizePaged<T>(raw: any, page: number, limit: number): PagedResponse<T> {
  // Backend hiện tại của bạn đang trả thẳng array => normalize thành dạng paged
  if (Array.isArray(raw)) {
    return {
      items: raw as T[],
      page,
      limit,
      has_more: (raw as T[]).length === limit,
    };
  }

  // Nếu sau này backend đổi sang trả paged object
  if (raw && typeof raw === "object" && Array.isArray(raw.items)) {
    return {
      items: raw.items as T[],
      page: Number(raw.page ?? page),
      limit: Number(raw.limit ?? limit),
      has_more: Boolean(raw.has_more),
    };
  }

  return { items: [], page, limit, has_more: false };
}

async function getPopular(page = 1, limit = 20): Promise<PagedResponse<Manga>> {
  // Route này chắc chắn tồn tại theo handler bạn gửi: /manga/popular?limit=
  const res = await http.get("/manga/popular", { params: { page, limit } });
  return normalizePaged<Manga>(res.data, page, limit);
}
const handleSearch = async () => {
  if (!query.trim()) return fetchPopular();
  setLoading(true);
  setError(null);
  try {
    const data = await mangaService.search(query); // dùng searchManga() bên trong
    setManga(data.results ?? []);
  } catch (err) {
    setError("Search failed. Please try again.");
    setManga([]);
    console.error(err);
  } finally {
    setLoading(false);
  }
};

async function search(query: string, page = 1, limit = 20): Promise<SearchResponse> {
  // QUAN TRỌNG: Đừng gọi "/manga/search" nếu backend chưa có route đó.
  // Dùng searchManga() vì nó đang trỏ đúng endpoint hiện có.
  // (page/limit tạm thời chưa dùng được cho search cho tới khi backend hỗ trợ)
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
  getPopular,   // trả về {items, page, limit, has_more}
  search,       // trả về SearchResponse (không 404)
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
