import { useContext } from "react";
import { TicketContext } from "../context/TicketContext";

export default function TicketPanel() {
  const { ticket } = useContext(TicketContext);

  return (
    <h2 style={{ fontSize: "2rem" }}>
      Remaining Tickets: {ticket}
    </h2>
  );
}