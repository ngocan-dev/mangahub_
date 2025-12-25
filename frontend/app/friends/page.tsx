"use client";

import { useEffect, useRef, useState } from "react";

import ProtectedRoute from "../components/ProtectedRoute";
import { friendService, type FriendRequest, type UserSummary } from "@/service/friend.service";

type DirectMessage = {
  from: number;
  to: number;
  message: string;
  timestamp: number;
};

export default function FriendsPage() {
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

  // ------------------------
  // Loaders
  // ------------------------

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

  const disconnectSocket = () => {
    if (reconnectRef.current) clearTimeout(reconnectRef.current);
    reconnectRef.current = null;

    if (socketRef.current) {
      socketRef.current.close();
      socketRef.current = null;
    }
  };

  const connectChat = (friend: UserSummary) => {
    disconnectSocket();
    setSelectedFriend(friend);
    setChatMessages([]);
    setChatError(null);

    const ws = new WebSocket(`ws://localhost:8080/ws/chat?friend_id=${friend.id}`);
    socketRef.current = ws;

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data) as DirectMessage;
        setChatMessages((prev) => [...prev, msg]);
      } catch (e) {
        console.error(e);
      }
    };

    ws.onclose = () => {
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

  useEffect(() => {
    void loadFriends();
    void loadPendingRequests();
    return () => disconnectSocket();
  }, []);

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
          {friends.map((f) => (
            <div key={f.id} className="flex justify-between">
              <span>{f.username}</span>
              <button onClick={() => connectChat(f)}>Chat</button>
            </div>
          ))}
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
