"use client";

import { useEffect, useMemo, useState } from "react";

import ProtectedRoute from "../components/ProtectedRoute";
import { analyticsService, type ReadingAnalyticsPoint, type ReadingStatistic } from "@/service/analytics.service"; // Adjusted path

export default function AnalyticsPage() {
  const [stats, setStats] = useState<ReadingStatistic | null>(null);
  const [analytics, setAnalytics] = useState<ReadingAnalyticsPoint[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadAnalytics = async () => {
      setLoading(true);
      setError(null);
      try {
        const [statData, analyticsData] = await Promise.all([
          analyticsService.getReadingStatistics(),
          analyticsService.getReadingAnalytics(),
        ]);
        setStats(statData);
        setAnalytics(analyticsData);
      } catch (err) {
        setError("Unable to load analytics right now.");
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    void loadAnalytics();
  }, []);

  const maxChapters = useMemo(
    () => (analytics.length ? Math.max(...analytics.map((item) => item.chaptersRead)) : 0),
    [analytics],
  );

  return (
    <ProtectedRoute>
      <section className="space-y-4">
        <div className="rounded-xl border border-slate-800 bg-slate-900 p-4 shadow-lg">
          <h1 className="text-2xl font-semibold text-white">Reading Analytics</h1>
          <p className="text-sm text-slate-400">Track your reading progress and streaks.</p>
        </div>

        {loading ? <p className="text-slate-300">Loading analytics...</p> : null}
        {error ? <p className="text-red-400">{error}</p> : null}

        {stats ? (
          <div className="grid gap-4 md:grid-cols-3">
            <div className="card">
              <p className="text-sm text-slate-400">Chapters read</p>
              <p className="text-2xl font-semibold text-white">{stats.totalChaptersRead}</p>
            </div>
            <div className="card">
              <p className="text-sm text-slate-400">Time spent</p>
              <p className="text-2xl font-semibold text-white">
                {stats.totalTimeMinutes ? `${stats.totalTimeMinutes} min` : "N/A"}
              </p>
            </div>
            <div className="card">
              <p className="text-sm text-slate-400">Streak</p>
              <p className="text-2xl font-semibold text-white">{stats.streakDays ?? 0} days</p>
            </div>
          </div>
        ) : null}

        {analytics.length ? (
          <div className="card">
            <h2 className="text-lg font-semibold text-white">Recent reading</h2>
            <div className="mt-4 grid grid-cols-2 gap-3 md:grid-cols-4">
              {analytics.map((point) => {
                const height = maxChapters ? Math.round((point.chaptersRead / maxChapters) * 100) : 0;
                return (
                  <div key={point.date} className="flex flex-col items-center gap-2 text-sm text-slate-300">
                    <div className="flex h-28 w-full items-end rounded-lg bg-slate-800 p-1">
                      <div className="w-full rounded bg-secondary" style={{ height: `${height}%` }} />
                    </div>
                    <span className="text-xs text-slate-500">{new Date(point.date).toLocaleDateString()}</span>
                    <span className="text-xs text-white">{point.chaptersRead} ch.</span>
                  </div>
                );
              })}
            </div>
          </div>
        ) : null}
      </section>
    </ProtectedRoute>
  );
}