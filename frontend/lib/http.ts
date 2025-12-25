import axios, {
  AxiosError,
  AxiosInstance,
  AxiosHeaders,
  InternalAxiosRequestConfig,
} from "axios";

import { ENV } from "@/config/env";

const getToken = (): string | null => {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("token");
};

const http: AxiosInstance = axios.create({
  baseURL: ENV.API_BASE_URL,
  timeout: 15000,
  headers: {
    "Content-Type": "application/json",
  },
});

http.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = getToken();

    if (token) {
      /**
       * Axios v1.x:
       * - config.headers là AxiosHeaders
       * - KHÔNG ĐƯỢC gán object thường
       * - CHỈ được set bằng .set()
       */
      if (!config.headers) {
        config.headers = new AxiosHeaders();
      }

      // Luôn dùng set – không branch linh tinh
      config.headers.set("Authorization", `Bearer ${token}`);
    }

    return config;
  },
  (error) => Promise.reject(error),
);

http.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401 && typeof window !== "undefined") {
      const code = (() => {
        const data = error.response?.data as
          | { code?: string; error?: string }
          | undefined;
        const value = data?.code ?? data?.error;
        return typeof value === "string" ? value.toUpperCase() : "";
      })();

      if (
        [
          "TOKEN_INVALID",
          "TOKEN_EXPIRED",
          "TOKEN_CLAIMS_INVALID",
          "TOKEN_NOT_BEFORE",
        ].includes(code)
      ) {
        localStorage.removeItem("token");
        window.dispatchEvent(new Event("auth:unauthorized"));
      }
    }

    return Promise.reject(error);
  },
);

export { http };
