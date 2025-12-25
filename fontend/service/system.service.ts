import { http } from "@/lib/http";

export interface ServerStatus {
  status: string;
  uptime?: string;
  version?: string;
  message?: string;
}

export interface SyncStatus {
  status: string;
  lastSync?: string;
  details?: string;
}

async function getServerStatus(): Promise<ServerStatus> {
  const { data } = await http.get<ServerStatus>("/server/status");
  return data;
}

async function getSyncStatus(): Promise<SyncStatus> {
  const { data } = await http.get<SyncStatus>("/sync/status");
  return data;
}

export const systemService = {
  getServerStatus,
  getSyncStatus,
};