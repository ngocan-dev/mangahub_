import { ENV } from "@/config/env";

const API_BASE_URL = ENV.API_BASE_URL;

const getAuthToken = (): string | null => {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("token");
};

interface ApiRequestOptions extends RequestInit {
  skipAuth?: boolean;
}

async function request<T>(path: string, options: ApiRequestOptions = {}): Promise<T> {
  const { skipAuth, headers, ...rest } = options;
  const token = skipAuth ? null : getAuthToken();
  const mergedHeaders: HeadersInit = {
    "Content-Type": "application/json",
    ...(headers ?? {}),
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...rest,
    headers: mergedHeaders,
  });

  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || response.statusText);
  }

  return (await response.json()) as T;
}

export interface User {
  id: number;
  username: string;
  email: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface Manga {
  id: number;
  title: string;
  description: string;
  cover_url: string;
  rating: number;
  views: number;
  status: "ongoing" | "completed";
  genres: string[];
}

export interface Chapter {
  id: number;
  manga_id: number;
  chapter_number: number;
  title: string;
  content_url: string;
}

export interface Review {
  id: number;
  manga_id?: number;
  rating: number;
  comment?: string;
}

export async function login(email: string, password: string): Promise<AuthResponse> {
  return request<AuthResponse>("/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
    skipAuth: true,
  });
}

export async function register(username: string, email: string, password: string): Promise<AuthResponse> {
  return request<AuthResponse>("/register", {
    method: "POST",
    body: JSON.stringify({ username, email, password }),
    skipAuth: true,
  });
}

export async function getPopularManga(): Promise<Manga[]> {
  return request<Manga[]>("/manga/popular", { method: "GET" });
}

export async function searchManga(keyword: string): Promise<Manga[]> {
  const query = new URLSearchParams({ q: keyword });
  return request<Manga[]>(`/manga/search?${query.toString()}`, { method: "GET" });
}

export async function getMangaById(id: number): Promise<Manga> {
  return request<Manga>(`/manga/${id}`, { method: "GET" });
}

export async function getChaptersByMangaId(id: number): Promise<Chapter[]> {
  return request<Chapter[]>(`/chapters/${id}`, { method: "GET" });
}

export async function getReviewsByMangaId(id: number): Promise<Review[]> {
  try {
    return await request<Review[]>(`/manga/${id}/reviews`, { method: "GET" });
  } catch (error) {
    console.error("Failed to load reviews", error);
    return [];
  }
}
