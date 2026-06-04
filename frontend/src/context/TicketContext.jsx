import { createContext, useState } from "react";

export const TicketContext = createContext();

export function TicketProvider({ children }) {

  // 剩餘票數
  const [ticket, setTicket] = 
  useState({
    1: 0, 
    2: 0, 
  });


  // 狀態文字
  const [status, setStatus] =
    useState("Ready to buy 🎯");

  // 即時事件 logs
  const [logs, setLogs] = useState([]);
  const [toasts, setToasts] =useState([]);

  // 區域選擇
  const [selectedArea, setSelectedArea] = useState(1)

  //

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

        // message
        toasts,
        setToasts,

        //area
        selectedArea,
        setSelectedArea
      }}
    >
      {children}
    </TicketContext.Provider>
  );
}