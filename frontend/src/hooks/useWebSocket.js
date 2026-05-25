import { useEffect } from "react";

export default function useWebSocket(onMessage) {
  useEffect(() => {
    const ws = new WebSocket("ws://localhost:8080/ws");

    ws.onopen = () => {
      console.log("✅ WebSocket connected");
    };

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      onMessage(data);
    };

    ws.onerror = (err) => {
      console.error("❌ WS error", err);
    };

    ws.onclose = () => {
      console.log("⚠️ WS closed");
    };

    return () => ws.close();
  }, []);
}