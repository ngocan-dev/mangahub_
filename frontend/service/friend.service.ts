import { http } from "@/lib/http";

export interface UserSummary {
  id: number;
  username: string;
}

export interface Activity {
  activity_id: number;
  user_id: number;
  username: string;
  activity_type: string;
  manga_id: number;
  manga_title?: string;
  manga_image?: string;
  rating?: number;
  review_id?: number;
  review_content?: string;
  completed_at?: string;
  created_at: string;
}

export interface ActivityFeedResponse {
  activities: Activity[];
  total: number;
  page: number;
  limit: number;
  pages: number;
}

async function searchUsers(query: string): Promise<UserSummary[]> {
  const { data } = await http.get<{ users: UserSummary[] }>("/users/search", { params: { query } });
  return data.users ?? [];
}

async function sendFriendRequest(username: string): Promise<void> {
  await http.post("/friends/requests", { username });
}

async function acceptFriendRequest(requesterUsername: string): Promise<void> {
  await http.post("/friends/requests/accept", { requester_username: requesterUsername });
}

async function getActivityFeed(page = 1, limit = 20): Promise<ActivityFeedResponse> {
  const { data } = await http.get<ActivityFeedResponse>("/friends/activity", { params: { page, limit } });
  return {
    activities: data.activities ?? [],
    total: data.total ?? 0,
    page: data.page ?? page,
    limit: data.limit ?? limit,
    pages: data.pages ?? 0,
  };
}

export const friendService = {
  searchUsers,
  sendFriendRequest,
  acceptFriendRequest,
  getActivityFeed,
};
