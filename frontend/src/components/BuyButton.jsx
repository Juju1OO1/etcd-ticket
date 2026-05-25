import axios from "axios";
import { useContext } from "react";
import { TicketContext } from "../context/TicketContext";

export default function BuyButton() {
  const { setStatus } = useContext(TicketContext);

  const handleBuy = async () => {
  setStatus("Processing...");

  try {
    const res = await axios.post("http://localhost:8080/buy");
    setStatus(res.data.message);
  } catch (err) {
    console.error(err);

    if (err.response) {
      setStatus(`Server Error: ${err.response.status}`);
    } else if (err.request) {
      setStatus("⚠️ Backend not running");
    } else {
      setStatus("❌ Unknown error");
    }
  }
};

  return (
    <button onClick={handleBuy} style={{ padding: "10px 20px" }}>
      Buy Ticket
    </button>
  );
}