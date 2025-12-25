"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { ENV } from "@/config/env";
import { useAuth } from "@/context/AuthContext";
import ProtectedRoute from "../components/ProtectedRoute";
import { friendService, type FriendRequest, type UserSummary } from "@/service/friend.service";

type DirectMessage = {
  from: number;
  to: number;
  message: string;
  timestamp: number;
};

export default function FriendsPage() {
  const { token } = useAuth();
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<UserSummary[]>([]);
  const [friends, setFriends] = useState<UserSummary[]>([]);
  const [pendingRequests, setPendingRequests] = useState<FriendRequest[]>([]);
  const [requestMessage, setRequestMessage] = useState<string | null>(null);
  const [requestId, setRequestId] = useState("");
  const [error, setError] = useState<string | null>(null);

  // Chat
  const [selectedFriend, setSelectedFriend] = useState<UserSummary | null>(null);
  const [chatMessages, setChatMessages] = useState<DirectMessage[]>([]);
  const [chatInput, setChatInput] = useState("");
  const [chatError, setChatError] = useState<string | null>(null);

  const socketRef = useRef<WebSocket | null>(null);
  const reconnectRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const presenceSocketRef = useRef<WebSocket | null>(null);
  const presenceReconnectRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [onlineUserIds, setOnlineUserIds] = useState<number[]>([]);

  // ------------------------
  // Loaders
  // ------------------------

  const loadFriends = useCallback(async () => {
    try {
      const data = await friendService.listFriends();
      setFriends(data);
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

  // ------------------------
  // Search
  // ------------------------

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

  // ------------------------
  // Accept / Reject (CÁCH A – CHUẨN REACT)
  // ------------------------

  const handleAcceptRequest = async (id: number) => {
    setError(null);
    try {
      await friendService.acceptFriendRequest(id);
      setRequestMessage("Request accepted.");
      setRequestId("");
      await loadFriends();
      await loadPendingRequests();
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

  // ------------------------
  // Chat (WS – chỉ dùng khi chat)
  // ------------------------
  const buildWSUrl = useCallback((path: string) => {
    const wsBase = ENV.API_BASE_URL.replace(/^http/i, "ws");
    const base = wsBase.endsWith("/") ? wsBase : `${wsBase}/`;
    return new URL(path, base);
  }, []);

  const disconnectSocket = useCallback(() => {
    if (reconnectRef.current) clearTimeout(reconnectRef.current);
    reconnectRef.current = null;

    if (socketRef.current) {
      socketRef.current.close();
      socketRef.current = null;
    }
  }, []);

  const connectChat = (friend: UserSummary) => {
    disconnectSocket();
    setSelectedFriend(friend);
    setChatMessages([]);
    setChatError(null);

    if (!token) {
      setChatError("Authentication required to start chat.");
      return;
    }

    const wsUrl = buildWSUrl("/ws/chat");
    wsUrl.searchParams.set("friend_id", String(friend.id));
    wsUrl.searchParams.set("token", token);

    const ws = new WebSocket(wsUrl.toString());
    socketRef.current = ws;

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data) as DirectMessage;
        setChatMessages((prev) => [...prev, msg]);
      } catch (e) {
        console.error(e);
      }
    };

    ws.onclose = (event) => {
      if (event.code === 1000 || !token) {
        return;
      }
      reconnectRef.current = setTimeout(() => connectChat(friend), 1500);
    };

    ws.onerror = () => {
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

  // ------------------------
  // Init
  // ------------------------
  const disconnectPresence = useCallback(() => {
    if (presenceReconnectRef.current) clearTimeout(presenceReconnectRef.current);
    presenceReconnectRef.current = null;

    if (presenceSocketRef.current) {
      presenceSocketRef.current.close();
      presenceSocketRef.current = null;
    }
  }, []);

  const connectPresence = useCallback(() => {
    disconnectPresence();

    if (!token) {
      setOnlineUserIds([]);
      return;
    }

    const wsUrl = buildWSUrl("/ws/chat");
    wsUrl.searchParams.set("token", token);

    const ws = new WebSocket(wsUrl.toString());
    presenceSocketRef.current = ws;

    ws.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data) as { type?: string; online_user_ids?: number[] };
        if (payload.type === "presence:update" && Array.isArray(payload.online_user_ids)) {
          setOnlineUserIds(payload.online_user_ids.map((id) => Number(id)).filter((id) => Number.isFinite(id)));
        }
      } catch (err) {
        console.error(err);
      }
    };

    ws.onclose = (event) => {
      if (event.code === 1000 || !token) return;
      presenceReconnectRef.current = setTimeout(() => connectPresence(), 1500);
    };

    ws.onerror = () => {
      setOnlineUserIds([]);
    };
  }, [buildWSUrl, disconnectPresence, token]);

  const onlineFriendIds = useMemo(() => new Set(onlineUserIds), [onlineUserIds]);
  const onlineFriends = useMemo(
    () => friends.filter((f) => onlineFriendIds.has(f.id)),
    [friends, onlineFriendIds],
  );

  useEffect(() => {
    void loadFriends();
    void loadPendingRequests();
    return () => {
      disconnectSocket();
      disconnectPresence();
    };
  }, [disconnectPresence, disconnectSocket, loadFriends, loadPendingRequests]);

  useEffect(() => {
    connectPresence();
    return () => disconnectPresence();
  }, [connectPresence, disconnectPresence]);

  // ------------------------
  // Render
  // ------------------------

  return (
    <ProtectedRoute>
      <section className="space-y-4">
        <h1 className="text-2xl font-semibold text-white">Friends</h1>

        {/* Search */}
        <div className="card space-y-2">
          <input
            placeholder="Search users"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleSearch()}
          />
          <button className="btn-primary" onClick={handleSearch}>
            Search
          </button>

          {results.map((user) => (
            <div key={user.id} className="flex justify-between">
              <span>{user.username}</span>
              <button onClick={() => handleSendRequest(user.id)}>Add</button>
            </div>
          ))}
        </div>

        {/* Pending */}
        <div className="card space-y-2">
          <input
            placeholder="Request ID"
            value={requestId}
            onChange={(e) => setRequestId(e.target.value)}
          />

          <button className="btn-primary w-full" onClick={handleAcceptByInput}>
            Accept request
          </button>

          <button className="btn-secondary w-full" onClick={handleRejectByInput}>
            Reject request
          </button>

          {pendingRequests.map((req) => (
            <div key={req.id} className="flex justify-between">
              <span>#{req.id} from {req.from_username}</span>
              <div className="flex gap-2">
                <button onClick={() => handleAcceptRequest(req.id)}>Accept</button>
                <button onClick={() => handleRejectRequest(req.id)}>Reject</button>
              </div>
            </div>
          ))}
        </div>

        {/* Friends */}
        <div className="card space-y-2">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">Friends</h2>
            <span className="text-sm text-gray-300">
              Online {onlineFriends.length}/{friends.length}
            </span>
          </div>
          {friends.map((f) => (
            <div key={f.id} className="flex justify-between">
              <div className="flex items-center gap-2">
                <span
                  aria-label={onlineFriendIds.has(f.id) ? "Online" : "Offline"}
                  className={`h-2 w-2 rounded-full ${onlineFriendIds.has(f.id) ? "bg-green-400" : "bg-gray-500"}`}
                />
                <span>{f.username}</span>
              </div>
              <button onClick={() => connectChat(f)}>Chat</button>
            </div>
          ))}
        </div>

        {/* Online friends */}
        <div className="card space-y-2">
          <h2 className="text-lg font-semibold">Online friends (WebSocket)</h2>
          {onlineFriends.length === 0 ? (
            <p className="text-sm text-gray-300">No friends online right now.</p>
          ) : (
            onlineFriends.map((f) => (
              <div key={f.id} className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="h-2 w-2 rounded-full bg-green-400" aria-hidden="true" />
                  <span>{f.username}</span>
                </div>
                <button onClick={() => connectChat(f)}>Chat</button>
              </div>
            ))
          )}
        </div>

        {/* Chat */}
        {selectedFriend && (
          <div className="card space-y-2">
            <h2>Chat with {selectedFriend.username}</h2>

            <div className="space-y-1">
              {chatMessages.map((m, i) => (
                <div key={i}>
                  <strong>{m.from === selectedFriend.id ? selectedFriend.username : "You"}:</strong>{" "}
                  {m.message}
                </div>
              ))}
            </div>

            <input
              value={chatInput}
              onChange={(e) => setChatInput(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && sendChatMessage()}
            />

            <button onClick={sendChatMessage}>Send</button>

            {chatError && <p className="text-red-400">{chatError}</p>}
          </div>
        )}

        {error && <p className="text-red-400">{error}</p>}
        {requestMessage && <p className="text-green-400">{requestMessage}</p>}
      </section>
    </ProtectedRoute>
  );
}
