"use client";

import { createContext, ReactNode, useCallback, useContext, useEffect, useMemo, useState } from "react";

import type { AuthResponse } from "@/service/auth.service";
import { authService } from "@/service/auth.service";

export interface AuthUser {
  id?: number;
  email: string;
  username?: string;
}

interface AuthContextValue {
  user: AuthUser | null;
  token: string | null;
  loading: boolean;
  isAuthenticated: boolean;
  login: (credentials: { email: string; password: string }) => Promise<void>;
  register: (payload: { username: string; email: string; password: string }) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState<boolean>(true);

  const persistToken = useCallback((value: string | null) => {
    if (typeof window === "undefined") return;
    if (value) {
      localStorage.setItem("token", value);
    } else {
      localStorage.removeItem("token");
    }
  }, []);

  const handleAuthSuccess = useCallback(
    (data: AuthResponse, fallbackEmail?: string) => {
      setToken(data.token);
      const derivedUser = data.user ?? (fallbackEmail ? { email: fallbackEmail } : null);
      setUser(derivedUser);
      persistToken(data.token);
    },
    [persistToken],
  );

  const logout = useCallback(() => {
    setUser(null);
    setToken(null);
    persistToken(null);
  }, [persistToken]);

  useEffect(() => {
    if (typeof window === "undefined") return;
    const storedToken = localStorage.getItem("token");
    if (storedToken) {
      setToken(storedToken);
      setUser((prev) => prev ?? null);
    }
    setLoading(false);
  }, []);

  useEffect(() => {
    const handleUnauthorized = () => logout();
    if (typeof window !== "undefined") {
      window.addEventListener("auth:unauthorized", handleUnauthorized);
    }
    return () => {
      if (typeof window !== "undefined") {
        window.removeEventListener("auth:unauthorized", handleUnauthorized);
      }
    };
  }, [logout]);

  const login = useCallback(
    async (credentials: { email: string; password: string }) => {
      setLoading(true);
      try {
        const data = await authService.login({ email: credentials.email, password: credentials.password });
        handleAuthSuccess(data, credentials.email);
      } finally {
        setLoading(false);
      }
    },
    [handleAuthSuccess],
  );

  const register = useCallback(
    async (payload: { username: string; email: string; password: string }) => {
      setLoading(true);
      try {
        const data = await authService.register({
          username: payload.username,
          email: payload.email,
          password: payload.password,
        });
        handleAuthSuccess(data, payload.email);
      } finally {
        setLoading(false);
      }
    },
    [handleAuthSuccess],
  );

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      token,
      loading,
      isAuthenticated: Boolean(token),
      login,
      register,
      logout,
    }),
    [loading, login, logout, register, token, user],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) throw new Error("useAuth must be used within AuthProvider");
  return context;
}
