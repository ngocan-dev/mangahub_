"use client";

import { useEffect, useState } from "react";

import { systemService, type ServerStatus, type SyncStatus } from "@/services/system.service";

type StatusType = "healthy" | "degraded" | "offline" | "unknown";

const statusColor: Record<StatusType, string> = {
  healthy: "bg-emerald-500/20 text-emerald-200 border-emerald-500/40",
  degraded: "bg-amber-500/20 text-amber-200 border-amber-500/40",
  offline: "bg-rose-500/20 text-rose-100 border-rose-500/40",
  unknown: "bg-slate-800 text-slate-200 border-slate-700",
};

const mapStatus = (value?: string): StatusType => {
  if (!value) return "unknown";
  const normalized = value.toLowerCase();
  if (["ok", "online", "healthy"].includes(normalized)) return "healthy";
  if (["degraded", "syncing", "pending"].includes(normalized)) return "degraded";
  if (["offline", "down"].includes(normalized)) return "offline";
  return "unknown";
};

export default function SystemStatusBanner() {
  const [serverStatus, setServerStatus] = useState<ServerStatus | null>(null);
  const [syncStatus, setSyncStatus] = useState<SyncStatus | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchStatus = async () => {
      setError(null);
      try {
        const [server, sync] = await Promise.all([
          systemService.getServerStatus(),
          systemService.getSyncStatus(),
        ]);
        setServerStatus(server);
        setSyncStatus(sync);
      } catch (err) {
        setError("Unable to contact MangaHub backend right now.");
        console.error(err);
      }
    };

    void fetchStatus();
    const interval = window.setInterval(fetchStatus, 60_000);
    return () => window.clearInterval(interval);
  }, []);

  if (error) {
    return (
      <div className="mb-4 rounded-lg border border-amber-500/40 bg-amber-500/10 px-4 py-3 text-sm text-amber-100">
        {error}
      </div>
    );
  }

  return (
    <div className="mb-4 grid gap-3 sm:grid-cols-2">
      <StatusCard
        label="API"
        status={mapStatus(serverStatus?.status)}
        primary={serverStatus?.status ?? "Unknown"}
        secondary={serverStatus?.message ?? `Uptime: ${serverStatus?.uptime ?? "n/a"}`}
      />
      <StatusCard
        label="Sync"
        status={mapStatus(syncStatus?.status)}
        primary={syncStatus?.status ?? "Unknown"}
        secondary={syncStatus?.details ?? `Last sync: ${syncStatus?.lastSync ?? "n/a"}`}
      />
    </div>
  );
}

function StatusCard({
  label,
  status,
  primary,
  secondary,
}: {
  label: string;
  status: StatusType;
  primary: string;
  secondary?: string;
}) {
  return (
    <div className={`flex flex-col rounded-lg border px-4 py-3 ${statusColor[status]}`}>
      <div className="flex items-center justify-between text-xs uppercase tracking-wide">
        <span>{label}</span>
        <span className="rounded-full bg-slate-950/50 px-2 py-0.5 text-[11px] font-semibold">{status}</span>
      </div>
      <p className="mt-2 text-base font-semibold leading-tight">{primary}</p>
      {secondary ? <p className="text-xs text-slate-200/80">{secondary}</p> : null}
    </div>
  );
}