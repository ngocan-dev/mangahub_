export const ENV = {
  API_BASE_URL: process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080",
} as const;

export type Environment = typeof ENV;
