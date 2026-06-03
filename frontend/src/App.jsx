import { useState } from "react";

import Home from "./pages/Home";
import Checkout from "./pages/Checkout";
import Success from "./pages/Success";
import UserInfo from "./pages/UserInfo";

import "./App.css";

function App() {

  const [page, setPage] =
    useState("landing");

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

  return (
    <Home setPage={setPage} />
  );
}

export default App;