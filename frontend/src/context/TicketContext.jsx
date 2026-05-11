import { createContext, useState } from "react";

export const TicketContext = createContext();

export function TicketProvider({ children }) {
  const [ticket, setTicket] = useState(10);
  const [status, setStatus] = useState("Ready to buy 🎯");

  return (
    <TicketContext.Provider
      value={{ ticket, setTicket, status, setStatus }}
    >
      {children}
    </TicketContext.Provider>
  );
}