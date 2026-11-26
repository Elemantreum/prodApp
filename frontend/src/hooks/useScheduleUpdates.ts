import { useEffect } from "react";

export function useScheduleUpdates(token: string | null, refresh: () => Promise<void>) {
  useEffect(() => {
    if (!token) return;
    const base = import.meta.env.VITE_API_BASE_URL ?? "/api/v1";
    const url = new URL(
      import.meta.env.VITE_WS_URL || base.replace(/\/$/, "") + "/schedule/live",
      location.href,
    );
    url.protocol = ["https:", "wss:"].includes(url.protocol) ? "wss:" : "ws:";
    let socket: WebSocket;
    let timer: ReturnType<typeof setTimeout>;
    let stopped = false;
    let retry = 1000;
    function connect() {
      socket = new WebSocket(url);
      socket.onopen = () => socket.send(token!);
      socket.onmessage = ({ data }) => {
        retry = 1000;
        if (data === "changed") void refresh();
      };
      socket.onclose = () => {
        if (stopped) return;
        timer = setTimeout(async () => {
          await refresh();
          if (!stopped) connect();
        }, retry);
        retry = Math.min(30000, retry * 2);
      };
    }
    connect();
    return () => {
      stopped = true;
      clearTimeout(timer);
      socket.onopen = null;
      socket.close();
    };
  }, [token, refresh]);
}
