import axios from "axios";

import { http } from "@/lib/http";

export interface UserSummary {
  id: number;
  username: string;
  email?: string;
  avatar?: string;
}

export interface FriendRequest {
  id: number;
  from_user_id: number;
  from_username?: string;
  to_user_id: number;
  status: string;
  created_at: string;
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
  payload?: Record<string, unknown>;
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
  try {
    const { data } = await http.get<{ users: UserSummary[] }>("/users/search", { params: { q: query } });
    return data?.users ?? [];
  } catch (error) {
    if (axios.isAxiosError(error)) {
      console.error("searchUsers failed", error.response?.data ?? error.message);
    } else {
      console.error("searchUsers failed", error);
    }
    throw error;
  }
}

async function sendFriendRequest(targetUserId: number): Promise<void> {
  try {
    await http.post("/friends/request", { target_user_id: targetUserId });
  } catch (error) {
    if (axios.isAxiosError(error)) {
      console.error("sendFriendRequest failed", error.response?.data ?? error.message);
    } else {
      console.error("sendFriendRequest failed", error);
    }
    throw error;
  }
}

async function acceptFriendRequest(requestId: number): Promise<void> {
  try {
    await http.post("/friends/accept", { request_id: requestId });
  } catch (error) {
    if (axios.isAxiosError(error)) {
      console.error("acceptFriendRequest failed", error.response?.data ?? error.message);
    } else {
      console.error("acceptFriendRequest failed", error);
    }
    throw error;
  }
}

async function rejectFriendRequest(requestId: number): Promise<void> {
  try {
    await http.post("/friends/reject", { request_id: requestId });
  } catch (error) {
    if (axios.isAxiosError(error)) {
      console.error("rejectFriendRequest failed", error.response?.data ?? error.message);
    } else {
      console.error("rejectFriendRequest failed", error);
    }
    throw error;
  }
}

async function listFriends(): Promise<UserSummary[]> {
  try {
    const { data } = await http.get<{ friends: UserSummary[] }>("/friends");
    return data?.friends ?? [];
  } catch (error) {
    if (axios.isAxiosError(error)) {
      console.error("listFriends failed", error.response?.data ?? error.message);
    } else {
      console.error("listFriends failed", error);
    }
    throw error;
  }
}

async function listPendingRequests(): Promise<FriendRequest[]> {
  try {
    const { data } = await http.get<{ requests: FriendRequest[] }>("/friends/requests");
    return data?.requests ?? [];
  } catch (error) {
    if (axios.isAxiosError(error)) {
      console.error("listPendingRequests failed", error.response?.data ?? error.message);
    } else {
      console.error("listPendingRequests failed", error);
    }
    throw error;
  }
}

async function getActivityFeed(page = 1, limit = 20): Promise<ActivityFeedResponse> {
  try {
    const { data } = await http.get<ActivityFeedResponse>("/friends/activity", { params: { page, limit } });
    return {
      activities: data?.activities ?? [],
      total: data?.total ?? 0,
      page: data?.page ?? page,
      limit: data?.limit ?? limit,
      pages: data?.pages ?? 0,
    };
  } catch (error) {
    if (axios.isAxiosError(error)) {
      console.error("getActivityFeed failed", error.response?.data ?? error.message);
    } else {
      console.error("getActivityFeed failed", error);
    }
    throw error;
  }
}

export const friendService = {
  searchUsers,
  sendFriendRequest,
  acceptFriendRequest,
  rejectFriendRequest,
  listFriends,
  listPendingRequests,
  getActivityFeed,
};
