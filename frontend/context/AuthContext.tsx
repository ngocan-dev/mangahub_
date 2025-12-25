"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";

import { authService, type AuthUser } from "@/service/auth.service";

/* =====================
 * Types
 * ===================== */

interface AuthContextValue {
  user: AuthUser | null;
  token: string | null;
  isAuthenticated: boolean;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (username: string, email: string, password: string) => Promise<void>;
  logout: () => void;
}

/* =====================
 * Context
 * ===================== */

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

/* =====================
 * Provider
 * ===================== */

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState<boolean>(true);

  /* ---------- helpers ---------- */

  const persistToken = (value: string | null) => {
    if (typeof window === "undefined") return;
    if (value) localStorage.setItem("token", value);
    else localStorage.removeItem("token");
  };

  /* ---------- derived ---------- */

  const isAuthenticated = useMemo(() => {
    return Boolean(user && token);
  }, [user, token]);

  /* ---------- actions ---------- */

  const login = useCallback(async (email: string, password: string) => {
    setLoading(true);
    const data = await authService.login({ email, password });

    setToken(data.token);
    setUser(data.user);
    persistToken(data.token);
    setLoading(false);
  }, []);

  const register = useCallback(
    async (username: string, email: string, password: string) => {
      setLoading(true);
      const data = await authService.register({
        username,
        email,
        password,
      });

      setToken(data.token);
      setUser(data.user);
      persistToken(data.token);
      setLoading(false);
    },
    [],
  );

  const logout = useCallback(() => {
    setUser(null);
    setToken(null);
    persistToken(null);
    setLoading(false);
  }, []);

  /* ---------- bootstrap ---------- */

  useEffect(() => {
    const bootstrap = async () => {
      const storedToken =
        typeof window !== "undefined" ? localStorage.getItem("token") : null;

      if (!storedToken) {
        setLoading(false);
        return;
      }

      try {
        setToken(storedToken);
        const me = await authService.getMe();
        setUser(me);
      } catch {
        logout();
        return;
      } finally {
        setLoading(false);
      }
    };

    void bootstrap();
  }, [logout]);

  /* ---------- render ---------- */

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        isAuthenticated,
        loading,
        login,
        register,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

/* =====================
 * Hook
 * ===================== */

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return ctx;
}
