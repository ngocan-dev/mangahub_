"use client";

import { useEffect, useMemo, useRef, useState } from "react";

import ProtectedRoute from "../components/ProtectedRoute"; // Adjusted path to match the file structure
import { friendService, type Activity, type FriendRequest, type UserSummary } from "@/service/friend.service";

type DirectMessage = {
  from: number;
  to: number;
  message: string;
  timestamp: number;
};

export default function FriendsPage() {
  const [query, setQuery] = useState<string>("");
  const [results, setResults] = useState<UserSummary[]>([]);
  const [activity, setActivity] = useState<Activity[]>([]);
  const [friends, setFriends] = useState<UserSummary[]>([]);
  const [pendingRequests, setPendingRequests] = useState<FriendRequest[]>([]);
  const [requestMessage, setRequestMessage] = useState<string | null>(null);
  const [requestId, setRequestId] = useState<string>("");
  const [error, setError] = useState<string | null>(null);
  const [chatError, setChatError] = useState<string | null>(null);
  const [selectedFriend, setSelectedFriend] = useState<UserSummary | null>(null);
  const [chatMessages, setChatMessages] = useState<DirectMessage[]>([]);
  const [chatInput, setChatInput] = useState<string>("");
  const socketRef = useRef<WebSocket | null>(null);
  const reconnectRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const wsBase = useMemo(() => {
    if (typeof window === "undefined") return "";
    const env = process.env.NEXT_PUBLIC_WS_URL;
    if (env) return env.replace(/^http/, "ws");
    return window.location.origin.replace(/^http/, "ws");
  }, []);

  const loadActivity = async () => {
    try {
      const feed = await friendService.getActivityFeed();
      setActivity(feed.activities);
    } catch (err) {
      setError("Could not load friend activity.");
      console.error(err);
    }
  };

  const loadFriends = async () => {
    try {
      const data = await friendService.listFriends();
      setFriends(data);
    } catch (err) {
      console.error(err);
    }
  };

  const loadPendingRequests = async () => {
    try {
      const data = await friendService.listPendingRequests();
      setPendingRequests(data);
    } catch (err) {
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

  const handleSendRequest = async (userId: number) => {
    setRequestMessage(null);
    try {
      await friendService.sendFriendRequest(userId);
      setRequestMessage("Friend request sent.");
    } catch (err) {
      setError("Could not send friend request.");
      console.error(err);
    }
  };

  const handleAcceptRequest = async (requestIdOverride?: number) => {
    const id = requestIdOverride ?? Number.parseInt(requestId.trim(), 10);
    if (!id) return;
    setError(null);
    try {
      await friendService.acceptFriendRequest(id);
      setRequestMessage("Request accepted.");
      setRequestId("");
      await loadActivity();
      await loadFriends();
      await loadPendingRequests();
    } catch (err) {
      setError("Unable to accept request.");
      console.error(err);
    }
  };

  const handleRejectRequest = async (requestIdOverride?: number) => {
    const id = requestIdOverride ?? Number.parseInt(requestId.trim(), 10);
    if (!id) return;
    setError(null);
    try {
      await friendService.rejectFriendRequest(id);
      setRequestMessage("Request rejected.");
      setRequestId("");
      await loadPendingRequests();
    } catch (err) {
      setError("Unable to reject request.");
      console.error(err);
    }
  };

  const disconnectSocket = () => {
    reconnectRef.current && clearTimeout(reconnectRef.current);
    reconnectRef.current = null;
    if (socketRef.current) {
      socketRef.current.close();
      socketRef.current = null;
    }
  };

  const connectChat = (friend: UserSummary) => {
    if (!wsBase) return;
    disconnectSocket();
    setSelectedFriend(friend);
    setChatMessages([]);
    setChatError(null);

    const ws = new WebSocket(`${wsBase}/ws/chat?friend_id=${friend.id}`);
    socketRef.current = ws;
    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data) as DirectMessage;
        setChatMessages((prev) => [...prev, msg]);
      } catch (e) {
        console.error("failed to parse message", e);
      }
    };
    ws.onclose = () => {
      reconnectRef.current = setTimeout(() => connectChat(friend), 1500);
    };
    ws.onerror = (evt) => {
      console.error("chat socket error", evt);
      setChatError("Chat connection error");
    };
  };

  const sendChatMessage = () => {
    if (!socketRef.current || !selectedFriend) return;
    const trimmed = chatInput.trim();
    if (!trimmed) return;
    try {
      socketRef.current.send(JSON.stringify({ message: trimmed }));
      setChatInput("");
    } catch (err) {
      setChatError("Unable to send message");
      console.error(err);
    }
  };

  useEffect(() => {
    void loadActivity();
    void loadFriends();
    void loadPendingRequests();
    return () => {
      disconnectSocket();
    };
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
                  <button className="btn-secondary text-xs" onClick={() => handleSendRequest(user.id)}>
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
                placeholder="Request ID"
                value={requestId}
                onChange={(e) => setRequestId(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && handleAcceptRequest()}
              />
              <button className="btn-primary w-full" onClick={handleAcceptRequest}>
                Accept request
              </button>
              <button className="btn-secondary w-full" onClick={handleRejectRequest}>
                Reject request
              </button>
              {requestMessage ? <p className="text-sm text-emerald-300">{requestMessage}</p> : null}
            </div>
            <div className="space-y-2 rounded-lg bg-slate-800 p-3">
              <p className="text-sm font-semibold text-white">Pending incoming</p>
              {pendingRequests.map((req) => (
                <div key={req.id} className="flex items-center justify-between text-sm text-slate-200">
                  <span>
                    Request #{req.id} from {req.from_username ?? `user ${req.from_user_id}`}
                  </span>
                  <div className="flex gap-2">
                    <button className="btn-secondary text-xs" onClick={() => handleAcceptRequest(req.id)}>
                      Accept
                    </button>
                    <button className="btn-secondary text-xs" onClick={() => handleRejectRequest(req.id)}>
                      Reject
                    </button>
                  </div>
                </div>
              ))}
              {!pendingRequests.length ? <p className="text-xs text-slate-400">No incoming requests.</p> : null}
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
                  <p className="text-slate-300">
                    {item.activity_type} {item.payload?.current_chapter ? `(Chapter ${item.payload.current_chapter as number})` : ""}
                  </p>
                  {item.manga_title ? <p className="text-xs text-slate-400">Manga: {item.manga_title}</p> : null}
                  {item.payload?.content ? (
                    <p className="text-xs text-slate-400">Review: {(item.payload.content as string).slice(0, 80)}</p>
                  ) : null}
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-sm text-slate-400">No activity yet.</p>
          )}
        </div>
        <div className="card space-y-3">
          <h2 className="text-lg font-semibold text-white">Friends</h2>
          <div className="space-y-2">
            {friends.map((friend) => (
              <div key={friend.id} className="flex items-center justify-between rounded-lg bg-slate-800 px-3 py-2">
                <div>
                  <p className="text-sm font-medium text-white">{friend.username}</p>
                  <p className="text-xs text-slate-400">{friend.email}</p>
                </div>
                <button className="btn-secondary text-xs" onClick={() => connectChat(friend)}>
                  Chat
                </button>
              </div>
            ))}
            {!friends.length ? <p className="text-sm text-slate-400">No friends yet.</p> : null}
          </div>
        </div>

        {selectedFriend ? (
          <div className="card space-y-3">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold text-white">Chat with {selectedFriend.username}</h2>
              <button className="btn-secondary" onClick={disconnectSocket}>
                Disconnect
              </button>
            </div>
            <div className="max-h-64 overflow-y-auto space-y-2 rounded-lg bg-slate-800 p-3">
              {chatMessages.map((msg, idx) => (
                <div key={`${msg.timestamp}-${idx}`} className="text-sm">
                  <span className="font-semibold text-emerald-300">{msg.from === selectedFriend.id ? selectedFriend.username : "You"}</span>:{" "}
                  <span>{msg.message}</span>
                </div>
              ))}
              {!chatMessages.length ? <p className="text-sm text-slate-400">No messages yet.</p> : null}
            </div>
            <div className="flex gap-2">
              <input
                className="w-full"
                placeholder="Say hi..."
                value={chatInput}
                onChange={(e) => setChatInput(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && sendChatMessage()}
              />
              <button className="btn-primary" onClick={sendChatMessage}>
                Send
              </button>
            </div>
            {chatError ? <p className="text-sm text-red-400">{chatError}</p> : null}
          </div>
        ) : null}

        {error ? <p className="text-red-400">{error}</p> : null}
      </section>
    </ProtectedRoute>
  );
}
