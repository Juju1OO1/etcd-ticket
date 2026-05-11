import { useContext } from "react";
import { TicketContext } from "../context/TicketContext";

export default function StatusBoard() {
  const { status } = useContext(TicketContext);

  return <p>Status: {status}</p>;
}