"use client";

import { useEffect, useState } from "react";

import ProtectedRoute from "../components/ProtectedRoute"; // Adjusted path to match the file structure
import { friendService, type Activity, type UserSummary } from "@/service/friend.service";

export default function FriendsPage() {
  const [query, setQuery] = useState<string>("");
  const [results, setResults] = useState<UserSummary[]>([]);
  const [activity, setActivity] = useState<Activity[]>([]);
  const [requestMessage, setRequestMessage] = useState<string | null>(null);
  const [requestUsername, setRequestUsername] = useState<string>("");
  const [error, setError] = useState<string | null>(null);

  const loadActivity = async () => {
    try {
      const feed = await friendService.getActivityFeed();
      setActivity(feed.activities);
    } catch (err) {
      setError("Could not load friend activity.");
      console.error(err);
    }
  };

  const handleSearch = async () => {
    setError(null);
    try {
      const data = await friendService.searchUsers(query);
      setResults(data);
    } catch (err) {
      setError("Unable to search users.");
      console.error(err);
    }
  };

  const handleSendRequest = async (username: string) => {
    setRequestMessage(null);
    try {
      await friendService.sendFriendRequest(username);
      setRequestMessage("Friend request sent.");
    } catch (err) {
      setError("Could not send friend request.");
      console.error(err);
    }
  };

  const handleAcceptRequest = async () => {
    if (!requestUsername.trim()) return;
    setError(null);
    try {
      await friendService.acceptFriendRequest(requestUsername.trim());
      setRequestMessage("Request accepted.");
      setRequestUsername("");
      await loadActivity();
    } catch (err) {
      setError("Unable to accept request.");
      console.error(err);
    }
  };

  useEffect(() => {
    void loadActivity();
  }, []);

  return (
    <ProtectedRoute>
      <section className="space-y-4">
        <div className="rounded-xl border border-slate-800 bg-slate-900 p-4 shadow-lg">
          <h1 className="text-2xl font-semibold text-white">Friends & Activity</h1>
          <p className="text-sm text-slate-400">
            Discover readers, send friend requests, and keep up with your friends&apos; activity.
          </p>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <div className="card space-y-3">
            <h2 className="text-lg font-semibold text-white">Find readers</h2>
            <div className="flex gap-2">
              <input
                className="w-full"
                placeholder="Search by username or email"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && handleSearch()}
              />
              <button className="btn-primary" onClick={handleSearch}>
                Search
              </button>
            </div>
            <div className="space-y-2">
              {results.map((user) => (
                <div key={user.id} className="flex items-center justify-between rounded-lg bg-slate-800 px-3 py-2">
                  <div>
                    <p className="text-sm font-medium text-white">{user.username ?? "Reader"}</p>
                    <p className="text-xs text-slate-400">{user.email}</p>
                  </div>
                  <button className="btn-secondary text-xs" onClick={() => handleSendRequest(user.username)}>
                    Add friend
                  </button>
                </div>
              ))}
              {!results.length ? <p className="text-sm text-slate-400">No users found yet.</p> : null}
            </div>
          </div>

          <div className="card space-y-3">
            <h2 className="text-lg font-semibold text-white">Respond to requests</h2>
            <div className="space-y-2">
              <input
                placeholder="Requester username"
                value={requestUsername}
                onChange={(e) => setRequestUsername(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && handleAcceptRequest()}
              />
              <button className="btn-primary w-full" onClick={handleAcceptRequest}>
                Accept request
              </button>
              {requestMessage ? <p className="text-sm text-emerald-300">{requestMessage}</p> : null}
            </div>
          </div>
        </div>

        <div className="card space-y-3">
          <h2 className="text-lg font-semibold text-white">Recent activity</h2>
          {activity.length ? (
            <ul className="space-y-2">
              {activity.map((item) => (
                <li key={item.activity_id} className="rounded-lg bg-slate-800 px-3 py-2 text-sm text-slate-200">
                  <div className="flex items-center justify-between">
                    <span>{item.username ?? "Someone"}</span>
                    <span className="text-xs text-slate-500">{new Date(item.created_at).toLocaleString()}</span>
                  </div>
                  <p className="text-slate-300">{item.activity_type}</p>
                  {item.manga_title ? <p className="text-xs text-slate-400">Manga: {item.manga_title}</p> : null}
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-sm text-slate-400">No activity yet.</p>
          )}
        </div>
        {error ? <p className="text-red-400">{error}</p> : null}
      </section>
    </ProtectedRoute>
  );
}
