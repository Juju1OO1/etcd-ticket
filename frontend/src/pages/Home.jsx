import { useContext } from "react";
import { TicketContext } from "../context/TicketContext";
import useWebSocket from "../hooks/useWebSocket";
import "./../App.css";

export default function Home() {
  const { ticket, setTicket, status, setStatus } = useContext(TicketContext);

  useWebSocket((data) => {
    if (data.type === "ticket_update") {
      setTicket(data.count);
    }

    if (data.type === "buy_result") {
      setStatus(data.message);
    }
  });

  const handleBuy = async () => {
    setStatus("Processing...");
  };

  const getStatusClass = () => {
    if (status.includes("Success")) return "success";
    if (status.includes("Sold")) return "error";
    if (status.includes("Processing")) return "warning";
    return "";
  };

  return (
    <div className="container">
      <div className="card">
        <h1 className="title">🎟️ Ticket System</h1>

        <div className="ticket">
          {ticket === 0 ? "SOLD OUT" : ticket}
        </div>
        
        <button
          className="button"
          disabled={ticket === 0}
          onClick={handleBuy}
          >
            {ticket === 0 ? "Sold Out" : "Buy Ticket"}
        </button>

        <div className={`status ${getStatusClass()}`}>
          {status}
        </div>
      </div>
    </div>
  );
}