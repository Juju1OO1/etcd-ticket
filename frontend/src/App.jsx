import { useState } from "react";
import Home from "./pages/Home";

function App() {
  const [start, setStart] = useState(false);

  if (start) {
    return <Home />;
  }

  return (
    <div style={{ textAlign: "center", marginTop: "100px" }}>
      <h1>🎟️ Ticket System Demo</h1>
      <button onClick={() => setStart(true)}>
        Start Demo
      </button>
    </div>
  );
}

export default App;