import { http } from "@/lib/http";

export interface ReadingStatistic {
  totalChaptersRead: number;
  totalTimeMinutes?: number;
  streakDays?: number;
}

export interface ReadingAnalyticsPoint {
  date: string;
  chaptersRead: number;
}

async function getReadingStatistics(): Promise<ReadingStatistic> {
  const { data } = await http.get<ReadingStatistic>("/statistics/reading");
  return data;
}

async function getReadingAnalytics(): Promise<ReadingAnalyticsPoint[]> {
  const { data } = await http.get<ReadingAnalyticsPoint[]>("/analytics/reading");
  return data;
}

export const analyticsService = {
  getReadingStatistics,
  getReadingAnalytics,
};
