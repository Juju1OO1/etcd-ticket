import { useContext } from "react";

import { TicketContext }
  from "../context/TicketContext";

export default function LogBoard() {

  const { logs } =
    useContext(TicketContext);

  return (

    <div className="logs">

      <h3>Realtime Events</h3>

      {logs.map((log, idx) => (

        <div key={idx}>
          {log}
        </div>

      ))}

    </div>
  );
}