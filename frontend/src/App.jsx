import { useState, useContext } from "react";
import { TicketContext } from "./context/TicketContext";
import ToastNotification from "./components/ToastNotification";
import Home from "./pages/Home";
import Checkout from "./pages/Checkout";
import Success from "./pages/Success";
import UserInfo from "./pages/UserInfo";
import useWebSocket from "./hooks/useWebSocket";



import "./App.css";

function App() {

  const [page, setPage] = useState("landing");
  const { 
    setTicket, 
    setStatus,
    toasts, 
    setToasts, 
    selectedArea
  } = useContext(TicketContext);
  

     useWebSocket((data) => {
    console.log("WS MESSAGE:", data);

    if (data.type === "ticket_available") {
      setTicket(prev => ({
        ...prev,
        [data.area_id]:
        data.available,
      }));

    if (
      data.area_id === selectedArea &&
      data.available === 0
    ) {
      setStatus("No Ticket 😔");
    }

}


    // distributed log
  if (data.type === "ticket_sold") {

  const id = Date.now();

  setToasts(prev => [
    ...prev,
    {
      id,
      user: data.user,
      area: data.area,
    },
  ]);

  setTimeout(() => {

    setToasts(prev =>
      prev.filter(
        t => t.id !== id
      )
    );

  }, 3000);
}


  });


  

  // ====================================
  // landing
  // ====================================

  if (page === "landing") { 

    return (
      <div className="landing">
        <h1>
          🎟️ Ticket System Demo
        </h1>
        <button
          className="button"
          onClick={() => setPage("userinfo")}
        >
          Enter
        </button>

      </div>
    );
  }

  // ====================================
  // checkout(Payment Info)
  // ====================================

  if (page === "checkout") {
    return (
    <Checkout setPage={setPage} />
  );
  }




  // ====================================
  // Payment Success
  // ====================================

  if (page === "success") {
    return (
    <Success setPage={setPage} />
  );
}

  // ====================================
  // UserInfo
  // ====================================

  if (page === "userinfo") {
    return (
      <UserInfo setPage={setPage} />
    );
  }


  // ====================================
  // home
  // ====================================

  if (page === "home") {
    return (
    <>
      <Home setPage={setPage} />

      <ToastNotification
        toasts={toasts}
      />
    </>
  );
}

// 最終 fallback
return <Home setPage={setPage} />;

}

export default App;