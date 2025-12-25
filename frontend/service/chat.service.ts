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
    const { data } = await http.get<{ conversations: Conversation[] }>(
      "/chat/conversations",
    );
    return data?.conversations ?? [];
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
