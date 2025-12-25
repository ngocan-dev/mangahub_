import { http } from "@/lib/http";

export interface AuthCredentials {
  email: string;
  password: string;
}

export interface RegisterPayload extends AuthCredentials {
  username: string;
}

export interface AuthUser {
  id: number;
  email: string;
  username?: string;
}

export interface AuthResponse {
  token: string;
  user?: AuthUser;
}

const normalizeUser = (user?: Partial<AuthUser> & Record<string, unknown>): AuthUser | undefined => {
  if (!user) return undefined;
  const idValue = typeof user.id === "number" ? user.id : Number(user.ID);
  const emailValue =
    typeof user.email === "string"
      ? user.email
      : typeof user.Email === "string"
        ? (user.Email as string)
        : "";
  const usernameValue =
    typeof user.username === "string"
      ? user.username
      : typeof user.Username === "string"
        ? (user.Username as string)
        : undefined;

  if (!Number.isFinite(idValue) || !emailValue) return undefined;
  return {
    id: idValue,
    email: emailValue,
    username: usernameValue,
  };
};

async function login(payload: AuthCredentials): Promise<AuthResponse> {
  const { data } = await http.post<AuthResponse>("/login", payload);
  return { token: data.token, user: normalizeUser(data.user) };
}

async function register(payload: RegisterPayload): Promise<AuthResponse> {
  const { data } = await http.post<AuthResponse>("/register", payload);
  return { token: data.token, user: normalizeUser(data.user) };
}

export const authService = { login, register };
