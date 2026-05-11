import { useContext } from "react";
import { TicketContext } from "../context/TicketContext";
import useWebSocket from "../hooks/useWebSocket";

import TicketPanel from "../components/TicketPanel";
import BuyButton from "../components/BuyButton";
import StatusBoard from "../components/StatusBoard";

export default function Home() {
  const { setTicket, setStatus } = useContext(TicketContext);

  useWebSocket((data) => {
    if (data.type === "ticket_update") {
      setTicket(data.count);
    }

    if (data.type === "buy_result") {
      setStatus(data.message);
    }
  });

  return (
    <div style={{ textAlign: "center", marginTop: "50px" }}>
      <h1>🎟️ Ticket System</h1>

      <TicketPanel />
      <BuyButton />
      <StatusBoard />
    </div>
  );
}