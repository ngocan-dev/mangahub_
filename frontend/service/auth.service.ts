import { http } from "@/lib/http";

/* =====================
 * Types
 * ===================== */

export interface AuthUser {
  id: number;
  username?: string;
  email?: string;
  avatar?: string;
}

export interface AuthCredentials {
  email: string;
  password: string;
}

export interface RegisterPayload {
  username: string;
  email: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  user: AuthUser;
}

/* =====================
 * Utils
 * ===================== */

/**
 * Normalize user object coming from API.
 * Accepts unknown / partial input safely.
 */
function normalizeUser(user: unknown): AuthUser {
  if (!user || typeof user !== "object") {
    throw new Error("Invalid user payload");
  }

  const u = user as Partial<AuthUser>;

  if (typeof u.id !== "number") {
    throw new Error("Invalid user fields");
  }

  return {
    id: u.id,
    username: typeof u.username === "string" ? u.username : undefined,
    email: typeof u.email === "string" ? u.email : undefined,
    avatar: typeof u.avatar === "string" ? u.avatar : undefined,
  };
}

/* =====================
 * API
 * ===================== */

async function login(payload: AuthCredentials): Promise<AuthResponse> {
  const { data } = await http.post<AuthResponse>("/login", payload);

  return {
    token: data.token,
    user: normalizeUser(data.user),
  };
}

async function register(payload: RegisterPayload): Promise<AuthResponse> {
  const { data } = await http.post<AuthResponse>("/register", payload);

  return {
    token: data.token,
    user: normalizeUser(data.user),
  };
}

async function getMe(): Promise<AuthUser> {
  const { data } = await http.get<AuthUser>("/me");
  return normalizeUser(data);
}

/* =====================
 * Export
 * ===================== */

export const authService = {
  login,
  register,
  getMe,
};
