import { http } from "@/lib/http";

export interface UserSummary {
  id: string;
  username?: string;
  email?: string;
}

export interface FriendActivity {
  id: string;
  action: string;
  createdAt: string;
  actor?: UserSummary;
}

export interface FriendRequestPayload {
  toUserId: string;
  message?: string;
}

async function searchUsers(query: string): Promise<UserSummary[]> {
  const { data } = await http.get<UserSummary[]>("/users/search", { params: { query } });
  return data;
}

async function sendFriendRequest(payload: FriendRequestPayload): Promise<void> {
  await http.post("/friends/requests", payload);
}

async function acceptFriendRequest(requestId: string): Promise<void> {
  await http.post("/friends/requests/accept", { requestId });
}

async function getActivityFeed(): Promise<FriendActivity[]> {
  const { data } = await http.get<FriendActivity[]>("/friends/activity");
  return data;
}

export const friendService = {
  searchUsers,
  sendFriendRequest,
  acceptFriendRequest,
  getActivityFeed,
};
