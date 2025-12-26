import axios from "axios";

import { http } from "@/lib/http";
import { type UserSummary } from "./friend.service";

export interface ChatRoom {
  id: number;
  code: string;
  name: string;
  is_private: boolean;
  created_by: number;
}

export interface ChatMessage {
  id: number;
  room_id: number;
  user_id: number;
  content: string;
  created_at: string;
}

export interface Conversation {
  room_id?: number;
  friend: UserSummary;
  last_message?: ChatMessage;
}

async function getConversations(): Promise<Conversation[]> {
  try {
    const { data } = await http.get<
      { conversations?: unknown } | Conversation[]
    >("/chat/conversations");

    const convs = Array.isArray(data)
      ? data
      : (data?.conversations as unknown);

    if (!Array.isArray(convs)) {
      return [];
    }

    return convs
      .map((item) => {
        const raw = item as Record<string, unknown>;
        const friendId = Number(raw.friend_id ?? (raw.friend as any)?.id);
        const friendUsername =
          typeof raw.friend_username === "string"
            ? raw.friend_username
            : (raw.friend as any)?.username;
        if (!Number.isFinite(friendId) || !friendUsername) return null;

        const friend: UserSummary = {
          id: friendId,
          username: friendUsername,
          avatar_url:
            typeof raw.friend_avatar === "string"
              ? raw.friend_avatar
              : (raw.friend as any)?.avatar_url,
        };

        const roomId = raw.room_id;
        const lastMessage = raw.last_message as ChatMessage | undefined;

        return {
          room_id: typeof roomId === "number" ? roomId : undefined,
          friend,
          last_message:
            lastMessage && typeof lastMessage === "object"
              ? {
                  ...lastMessage,
                  created_at: (lastMessage as any).created_at,
                }
              : undefined,
        } satisfies Conversation;
      })
      .filter(Boolean) as Conversation[];
  } catch (error) {
    if (axios.isAxiosError(error)) {
      console.error(
        "getConversations failed",
        error.response?.data ?? error.message,
      );
    } else {
      console.error("getConversations failed", error);
    }
    throw error;
  }
}

async function getRoomMessages(roomId: number): Promise<ChatMessage[]> {
  try {
    const { data } = await http.get<{ messages: ChatMessage[] }>(
      `/chat/rooms/${roomId}/messages`,
    );
    return data?.messages ?? [];
  } catch (error) {
    if (axios.isAxiosError(error)) {
      console.error(
        "getRoomMessages failed",
        error.response?.data ?? error.message,
      );
    } else {
      console.error("getRoomMessages failed", error);
    }
    throw error;
  }
}

async function sendMessage(
  friendUserId: number,
  content: string,
): Promise<{ room: ChatRoom; message: ChatMessage }> {
  try {
    const { data } = await http.post<{ room: ChatRoom; message: ChatMessage }>(
      "/chat/messages",
      {
        friend_user_id: friendUserId,
        content,
      },
    );
    return data;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      console.error(
        "sendMessage failed",
        error.response?.data ?? error.message,
      );
    } else {
      console.error("sendMessage failed", error);
    }
    throw error;
  }
}

export const chatService = {
  getConversations,
  getRoomMessages,
  sendMessage,
};
