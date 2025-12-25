import { http } from "@/lib/http";

export interface AuthCredentials {
  email: string;
  password: string;
}

export interface RegisterPayload extends AuthCredentials {
  username?: string;
}

export interface AuthResponse {
  token: string;
  user?: {
    id: string;
    email: string;
    username?: string;
  };
}

async function login(payload: AuthCredentials): Promise<AuthResponse> {
  const { data } = await http.post<AuthResponse>("/login", payload);
  return data;
}

async function register(payload: RegisterPayload): Promise<AuthResponse> {
  const { data } = await http.post<AuthResponse>("/register", payload);
  return data;
}

export const authService = { login, register };
