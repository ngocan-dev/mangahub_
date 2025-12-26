"use client";

import { useCallback, useEffect, useMemo, useState } from "react";

import ProtectedRoute from "../components/ProtectedRoute";
import { ENV } from "@/config/env";
import { useAuth } from "@/context/AuthContext";
import { chatService, type ChatMessage, type Conversation } from "@/service/chat.service";
import { friendService, type FriendRequest, type UserSummary } from "@/service/friend.service";

export default function FriendsPage() {
  const { user, token } = useAuth();
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<UserSummary[]>([]);
  const [friends, setFriends] = useState<UserSummary[]>([]);
  const [pendingRequests, setPendingRequests] = useState<FriendRequest[]>([]);
  const [requestMessage, setRequestMessage] = useState<string | null>(null);
  const [requestId, setRequestId] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [conversations, setConversations] = useState<Conversation[]>([]);

  const [selectedFriend, setSelectedFriend] = useState<UserSummary | null>(null);
  const [selectedRoomId, setSelectedRoomId] = useState<number | null>(null);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [chatInput, setChatInput] = useState("");
  const [chatError, setChatError] = useState<string | null>(null);
  const [onlineUserIds, setOnlineUserIds] = useState<Set<number>>(new Set());
  const [rawOnlineUserIds, setRawOnlineUserIds] = useState<Set<number>>(new Set());

  /* ------------------------ */
  /* Loaders */
  /* ------------------------ */

  const loadFriends = useCallback(async () => {
    try {
      const data = await friendService.listFriends();
      const unique = Array.from(new Map(data.map((f) => [f.id, f])).values());
      setFriends(unique);
    } catch (err) {
      console.error(err);
    }
  }, []);

  const loadPendingRequests = useCallback(async () => {
    try {
      const data = await friendService.listPendingRequests();
      setPendingRequests(data);
    } catch (err) {
      console.error(err);
    }
  }, []);

  const loadConversations = useCallback(async () => {
    try {
      const data = await chatService.getConversations();
      setConversations(data);
    } catch (err) {
      console.error(err);
    }
  }, []);

  /* ------------------------ */
  /* Derived data */
  /* ------------------------ */

  const roomByFriend = useMemo(() => {
    const map = new Map<number, number>();
    conversations.forEach((conv) => {
      if (conv.room_id) map.set(conv.friend.id, conv.room_id);
    });
    return map;
  }, [conversations]);

  const lastMessageByFriend = useMemo(() => {
    const map = new Map<number, ChatMessage | undefined>();
    conversations.forEach((conv) => {
      map.set(conv.friend.id, conv.last_message);
    });
    return map;
  }, [conversations]);

  const acceptedFriends = useMemo(() => {
    const map = new Map<number, UserSummary>();
    friends.forEach((f) => map.set(f.id, f));
    conversations.forEach((c) => map.set(c.friend.id, c.friend));
    return Array.from(map.values()).filter((f) => f.id !== user?.id);
  }, [conversations, friends, user?.id]);

  const onlineFriendsCount = useMemo(
    () => acceptedFriends.filter((f) => onlineUserIds.has(f.id)).length,
    [acceptedFriends, onlineUserIds],
  );

  /* ------------------------ */
  /* Search */
  /* ------------------------ */

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
      await loadPendingRequests();
      await loadFriends();
      setRequestMessage("Friend request sent.");
    } catch (err) {
      setError("Could not send friend request.");
      console.error(err);
    }
  };

  /* ------------------------ */
  /* Accept / Reject */
  /* ------------------------ */

  const handleAcceptRequest = async (id: number) => {
    setError(null);
    try {
      await friendService.acceptFriendRequest(id);
      setRequestMessage("Request accepted.");
      setRequestId("");
      await loadFriends();
      await loadPendingRequests();
      await loadConversations();
    } catch (err) {
      setError("Unable to accept request.");
      console.error(err);
    }
  };

  const handleRejectRequest = async (id: number) => {
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

  const handleAcceptByInput = async () => {
    const id = Number.parseInt(requestId.trim(), 10);
    if (!id) return;
    await handleAcceptRequest(id);
  };

  const handleRejectByInput = async () => {
    const id = Number.parseInt(requestId.trim(), 10);
    if (!id) return;
    await handleRejectRequest(id);
  };

  /* ------------------------ */
  /* Chat */
  /* ------------------------ */

  const selectFriend = useCallback(
    async (friend: UserSummary) => {
      setSelectedFriend(friend);
      setChatError(null);
      const roomId = roomByFriend.get(friend.id) ?? null;
      setSelectedRoomId(roomId);

      if (roomId) {
        try {
          const msgs = await chatService.getRoomMessages(roomId);
          setChatMessages(msgs);
        } catch (err) {
          setChatError("Unable to load messages");
          console.error(err);
          setChatMessages([]);
        }
      } else {
        setChatMessages([]);
      }
    },
    [roomByFriend],
  );

  const handleSendChatMessage = async () => {
    if (!selectedFriend) {
      setChatError("Select a friend to chat with.");
      return;
    }

    const trimmed = chatInput.trim();
    if (!trimmed) return;

    try {
      const { room, message } = await chatService.sendMessage(
        selectedFriend.id,
        trimmed,
      );
      setSelectedRoomId(room.id);
      setChatMessages((prev) => [...prev, message]);
      setChatInput("");
      setConversations((prev) => {
        const updated = [...prev];
        const idx = updated.findIndex((c) => c.friend.id === selectedFriend.id);
        const nextConv: Conversation = {
          friend: selectedFriend,
          room_id: room.id,
          last_message: message,
        };
        if (idx >= 0) {
          updated[idx] = { ...updated[idx], ...nextConv };
        } else {
          updated.unshift(nextConv);
        }
        return updated;
      });
    } catch (err) {
      setChatError("Unable to send message");
      console.error(err);
    }
  };

  /* ------------------------ */
  /* Init */
  /* ------------------------ */
  useEffect(() => {
    void loadFriends();
    void loadPendingRequests();
    void loadConversations();
  }, [loadConversations, loadFriends, loadPendingRequests]);

  useEffect(() => {
    if (!token) return;

    let socket: WebSocket | null = null;
    let reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
    let stopped = false;

    const buildPresenceWsUrl = (): string | null => {
      try {
        const apiBase = new URL(ENV.API_BASE_URL);
        const protocol = apiBase.protocol === "https:" ? "wss:" : "ws:";
        return `${protocol}//${apiBase.host}/ws/chat?token=${encodeURIComponent(token)}`;
      } catch (err) {
        console.error("Invalid API base URL for WebSocket", err);
        if (typeof window !== "undefined") {
          const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
          return `${protocol}//${window.location.host}/ws/chat?token=${encodeURIComponent(token)}`;
        }
        return null;
      }
    };

    const handlePresenceUpdate = (raw: string) => {
      try {
        const payload = JSON.parse(raw) as { type?: string; online_user_ids?: Array<number | string> };
        if (payload.type === "presence:update" && Array.isArray(payload.online_user_ids)) {
          const normalizedIds = payload.online_user_ids
            .map((id) => Number(id))
            .filter((id) => Number.isFinite(id));
          setRawOnlineUserIds(new Set(normalizedIds));
        }
      } catch (err) {
        console.error("Failed to parse presence payload", err);
      }
    };

    const connect = () => {
      const wsUrl = buildPresenceWsUrl();
      if (!wsUrl) return;

      try {
        socket = new WebSocket(wsUrl);
      } catch (err) {
        console.error("Failed to create presence WebSocket", err);
        if (!stopped) {
          reconnectTimeout = setTimeout(connect, 3000);
        }
        return;
      }

      socket.onmessage = (event) => {
        if (typeof event.data === "string") {
          handlePresenceUpdate(event.data);
        }
      };

      socket.onclose = () => {
        if (stopped) return;
        setOnlineUserIds(new Set());
        reconnectTimeout = setTimeout(connect, 3000);
      };

      socket.onerror = () => {
        socket?.close();
      };
    };

    connect();

    return () => {
      stopped = true;
      if (reconnectTimeout) clearTimeout(reconnectTimeout);
      socket?.close();
    };
  }, [token]);

  // Filter online ids to friends only to avoid leaking presence
  useEffect(() => {
    const friendIds = new Set(acceptedFriends.map((f) => f.id));
    const filtered = new Set(
      Array.from(rawOnlineUserIds).filter((id) => friendIds.has(id)),
    );
    setOnlineUserIds(filtered);
  }, [acceptedFriends, rawOnlineUserIds]);

  /* ------------------------ */
  /* Render helpers */
  /* ------------------------ */

  const renderLastMessage = (friend: UserSummary) => {
    const last = lastMessageByFriend.get(friend.id);
    if (!last) return <span className="text-xs text-gray-400">No messages yet</span>;
    const authorLabel = last.user_id === friend.id ? friend.username : "You";
    return (
      <span className="truncate text-xs text-gray-300">
        {authorLabel}: {last.content}
      </span>
    );
  };

  return (
    <ProtectedRoute>
      <section className="space-y-4">
        <h1 className="text-2xl font-semibold text-white">Friends</h1>

        <div className="grid items-start gap-4 lg:grid-cols-[340px_1fr]">
          {/* Left column */}
          <div className="space-y-4">
            {/* Search */}
            <div className="card space-y-3">
              <div className="flex items-center gap-2">
                <input
                  placeholder="Search users"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handleSearch()}
                />
                <button className="btn-primary" onClick={handleSearch}>
                  Search
                </button>
              </div>

              {results.map((user) => (
                <div key={user.id} className="flex items-center justify-between rounded bg-gray-800/40 px-3 py-2">
                  <div>
                    <p className="font-medium text-white">{user.username}</p>
                    <p className="text-xs text-gray-400">{user.email}</p>
                  </div>
                  <button className="btn-secondary" onClick={() => handleSendRequest(user.id)}>
                    Add
                  </button>
                </div>
              ))}
            </div>

            {/* Pending */}
            <div className="card space-y-3">
              <h2 className="text-lg font-semibold">Pending requests</h2>
              <div className="flex gap-2">
                <input
                  placeholder="Request ID"
                  value={requestId}
                  onChange={(e) => setRequestId(e.target.value)}
                />
                <button className="btn-primary w-24" onClick={handleAcceptByInput}>
                  Accept
                </button>
                <button className="btn-secondary w-24" onClick={handleRejectByInput}>
                  Reject
                </button>
              </div>

              {pendingRequests.length === 0 ? (
                <p className="text-sm text-gray-400">No pending requests.</p>
              ) : (
                pendingRequests.map((req) => (
                  <div key={req.id} className="flex items-center justify-between rounded bg-gray-800/40 px-3 py-2">
                    <div>
                      <p className="font-medium text-white">#{req.id} from {req.from_username}</p>
                      <p className="text-xs text-gray-400">{new Date(req.created_at).toLocaleString()}</p>
                    </div>
                    <div className="flex gap-2">
                      <button className="btn-primary" onClick={() => handleAcceptRequest(req.id)}>
                        Accept
                      </button>
                      <button className="btn-secondary" onClick={() => handleRejectRequest(req.id)}>
                        Reject
                      </button>
                    </div>
                  </div>
                ))
              )}
            </div>

            {/* Friends */}
            <div className="card space-y-3">
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-semibold">Friends</h2>
                <span className="text-xs text-gray-400">
                  Online {onlineFriendsCount}/{acceptedFriends.length}
                </span>
              </div>
              {acceptedFriends.length === 0 ? (
                <p className="text-sm text-gray-400">No friends yet.</p>
              ) : (
                <div className="space-y-2">
                  {acceptedFriends.map((f) => (
                    <button
                      key={f.id}
                      className={`w-full rounded px-3 py-2 text-left transition ${
                        selectedFriend?.id === f.id ? "bg-indigo-600/60" : "bg-gray-800/60 hover:bg-gray-700/60"
                      }`}
                      onClick={() => void selectFriend(f)}
                    >
                      <div className="flex items-center gap-2">
                        {onlineUserIds.has(f.id) && (
                          <span className="h-2 w-2 rounded-full bg-green-500" aria-label="Online indicator" />
                        )}
                        <p className="font-semibold text-white">{f.username}</p>
                      </div>
                      {renderLastMessage(f)}
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Right column */}
          <div className="space-y-4">
            <div className="card flex h-full min-h-[520px] flex-col space-y-3">
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-semibold">
                  {selectedFriend ? `Chat with ${selectedFriend.username}` : "Chat"}
                </h2>
              </div>

              {!selectedFriend ? (
                <p className="text-sm text-gray-400">Select a friend to start chatting.</p>
              ) : (
                <>
                  {selectedRoomId === null && (
                    <p className="text-sm text-gray-400">
                      No messages yet. Say hello to {selectedFriend.username}.
                    </p>
                  )}

                  <div className="flex-1 space-y-2 overflow-y-auto rounded bg-gray-900/40 p-3">
                    {chatMessages.map((m) => (
                      <div
                        key={m.id}
                        className={`flex ${m.user_id === user?.id ? "justify-end" : "justify-start"}`}
                      >
                        <div
                          className={`max-w-[75%] rounded px-3 py-2 text-sm ${
                            m.user_id === user?.id ? "bg-indigo-600 text-white" : "bg-gray-800 text-gray-100"
                          }`}
                        >
                          <p className="text-xs text-gray-300">
                            {m.user_id === user?.id ? "You" : selectedFriend.username} ·{" "}
                            {new Date(m.created_at).toLocaleString()}
                          </p>
                          <p className="mt-1 whitespace-pre-wrap break-words">{m.content}</p>
                        </div>
                      </div>
                    ))}
                  </div>

                  <div className="flex items-center gap-2">
                    <input
                      value={chatInput}
                      onChange={(e) => setChatInput(e.target.value)}
                      onKeyDown={(e) => e.key === "Enter" && handleSendChatMessage()}
                      placeholder="Type a message"
                    />
                    <button className="btn-primary" onClick={handleSendChatMessage}>
                      Send
                    </button>
                  </div>

                  {chatError && <p className="text-sm text-red-400">{chatError}</p>}
                </>
              )}
            </div>
          </div>
        </div>

        {error && <p className="text-red-400">{error}</p>}
        {requestMessage && <p className="text-green-400">{requestMessage}</p>}
      </section>
    </ProtectedRoute>
  );
}
