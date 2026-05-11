import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import { TicketProvider } from "./context/TicketContext";

ReactDOM.createRoot(document.getElementById("root")).render(
  <React.StrictMode>
    <TicketProvider>
      <App />
    </TicketProvider>
  </React.StrictMode>
);