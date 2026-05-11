import { useState } from "react";
import Home from "./pages/Home";
import "./App.css";

function App() {
  const [start, setStart] = useState(false);

  if (start) {
    return <Home />;
  }

  return (
    <div style={{ textAlign: "center", marginTop: "100px" }}>
      <h1>🎟️ Ticket System Demo</h1>
      <button className="button" onClick={() => setStart(true)}>
        Enter
      </button>
    </div>
  );
}

export default App;