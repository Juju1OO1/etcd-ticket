import { useContext } from "react";
import { TicketContext } from "../context/TicketContext";
import useWebSocket from "../hooks/useWebSocket";
import "./../App.css";

export default function Home({setPage}) {
  const { ticket, setTicket, status, setStatus, logs, setLogs, selectedArea, setSelectedArea } = useContext(TicketContext);
  
  // 上一頁 UserInfo 的資訊暫存在 local
  const userName = localStorage.getItem("userName");
  const phoneNum = localStorage.getItem("phoneNum");


  useWebSocket((data) => {
    if (data.type === "ticket_available") {
      setTicket(data.count);
    }

    if (data.type === "ticket_sold") {
      setStatus(data.message);
    }

    // distributed log
    if (data.type === "sold_log") {

      setLogs((prev) => [

        `🔥 ${data.user} bought Area-${data.area} ticket`,

        ...prev,
      ]);
    }



  });

  const handleBuy = async () => {

    setStatus("Processing...");

    try {

      const res = await fetch(
        "http://localhost:8080/api/tickets/reserve",
        {
          method: "POST",

          headers: {
            "Content-Type":
              "application/json",
          },

          body: JSON.stringify({

            user_name: userName,

            phone_num: phoneNum,

            area: selectedArea,

          }),
        }
      );

    const data = await res.json();

    console.log(data);

    setStatus(
      data.message || "Reserved"
    );

    // 搶票成功
    if (data.data?.reserved) {

      localStorage.setItem(
        "selectedArea",
        selectedArea
      );

      // 進付款頁
      setPage("checkout");
    }

  } catch {

    setStatus(
      "❌ Backend Error"
    );
  }
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
        <h1 className="title">
          🎟️ World Tour
        </h1>

        {/* LIVE EVENTS */}
        <div className="logs">

          <h3>Realtime Events</h3>

          {logs.map((log, idx) => (

            <div key={idx} className="log-item">
              {log}
            </div>

          ))}

        </div>

        <div className="area-selector">

          <h3>Select Area</h3>

          <select
            className="area-select"
            value={selectedArea}
            onChange={(e) =>
              setSelectedArea(Number(e.target.value))
            }
          >
            <option value={1}>
              Area 1
            </option>

            <option value={2}>
              Area 2
            </option>

          </select>

        </div>
 


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

        {/* realtime logs */}
        



      </div>
    </div>
  );
}