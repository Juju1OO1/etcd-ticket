import { useContext } from "react";

import { TicketContext }
  from "../context/TicketContext";

export default function AreaBlock({ area }) {

  const {
    tickets,
    status,
    setStatus,
  } = useContext(TicketContext);

  const ticket = tickets[area];

  // ====================================
  // buy
  // ====================================

  const handleBuy = async () => {

    setStatus(
      `Processing Area-${area}...`
    );

    try {

      await fetch(
        "http://localhost:8080/buy",
        {
          method: "POST",

          headers: {
            "Content-Type":
              "application/json",
          },

          body: JSON.stringify({

            userName: "john",

            phoneNum: "0912345678",

            area: area,

          }),
        }
      );

    } catch {

      setStatus(
        "❌ Backend error"
      );
    }
  };

  return (

    <div className="card">

      <h2>
        Area {area}
      </h2>

      <div className="ticket">

        {ticket === 0
          ? "SOLD OUT"
          : ticket}

      </div>

      <button
        className="button"

        disabled={ticket === 0}

        onClick={handleBuy}
      >

        {ticket === 0
          ? "Sold Out"
          : "Buy Ticket"}

      </button>

      <div className="status">
        {status}
      </div>

    </div>
  );
}