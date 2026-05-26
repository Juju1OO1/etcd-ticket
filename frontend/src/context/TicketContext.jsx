import { createContext, useState } from "react";

export const TicketContext = createContext();

export function TicketProvider({ children }) {

  // 剩餘票數
  const [ticket, setTicket] = useState(10);

  // 狀態文字
  const [status, setStatus] =
    useState("Ready to buy 🎯");

  // 即時事件 logs
  const [logs, setLogs] = useState([]);

  return (
    <TicketContext.Provider
      value={{

        // ticket
        ticket,
        setTicket,

        // status
        status,
        setStatus,

        // logs
        logs,
        setLogs,
      }}
    >
      {children}
    </TicketContext.Provider>
  );
}