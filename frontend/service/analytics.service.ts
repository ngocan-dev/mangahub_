import { http } from "@/lib/http";

export interface ReadingSummary {
  total_manga: number;
  total_chapters_read: number;
  reading_streak: number;
  last_read_at?: string;
}

export interface ReadingAnalyticsPoint {
  date: string;
  chapters_read: number;
}

export interface ReadingAnalyticsResponse {
  daily: ReadingAnalyticsPoint[];
  weekly: ReadingAnalyticsPoint[];
  monthly: ReadingAnalyticsPoint[];
}

async function getReadingStatistics(): Promise<ReadingSummary> {
  const { data } = await http.get<ReadingSummary>("/statistics/reading");
  return data;
}

async function getReadingAnalytics(): Promise<ReadingAnalyticsResponse> {
  const { data } = await http.get<ReadingAnalyticsResponse>("/analytics/reading");
  return data;
}

export const analyticsService = {
  getReadingStatistics,
  getReadingAnalytics,
};
