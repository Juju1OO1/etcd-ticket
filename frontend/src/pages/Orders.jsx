import { useState, useEffect } from "react";
import "./../App.css";

export default function Orders({ setPage }) {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const userName = localStorage.getItem("userName");
  const phoneNum = localStorage.getItem("phoneNum");

  useEffect(() => {
    if (!userName) {
      setError("No user info found.");
      setLoading(false);
      return;
    }

    fetch(`http://localhost:8080/api/orders?user_name=${encodeURIComponent(userName)}`)
      .then((res) => res.json())
      .then((data) => {
        if (data.data) {
          setOrders(data.data);
        } else {
          setError(data.message || "Failed to load orders.");
        }
      })
      .catch((err) => { console.error("Orders fetch error:", err); setError("Backend error: " + err.message); })
      .finally(() => setLoading(false));
  }, [userName]);

  const areaLabel = (area) => `Area ${area}`;

  const formatDate = (iso) =>
    new Date(iso).toLocaleString("zh-TW", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    });

  return (
    <div className="orders-container">
      <div className="orders-card">
        <h1>🗒️ My Orders</h1>

        <div className="order-item" style={{ marginBottom: "4px" }}>
          <div className="order-row">
            <span className="order-label">Name</span>
            <span className="order-value">{userName}</span>
          </div>
          <div className="order-row">
            <span className="order-label">Telephone</span>
            <span className="order-value">{phoneNum}</span>
          </div>
        </div>

        {loading && <p className="status">Loading...</p>}

        {error && <p className="status" style={{ color: "#ff4757" }}>{error}</p>}

        {!loading && !error && orders.length === 0 && (
          <p className="status">No orders found.</p>
        )}

        {!loading && !error && orders.length > 0 && (
          <div className="orders-list">
            {orders.map((o) => (
              <div key={o.OrderID} className="order-item">
                <div className="order-row">
                  <span className="order-label">Order ID</span>
                  <span className="order-value order-id">{o.OrderID}</span>
                </div>
                <div className="order-row">
                  <span className="order-label">Area</span>
                  <span className="order-value">{areaLabel(o.Area)}</span>
                </div>
                <div className="order-row">
                  <span className="order-label">Status</span>
                  <span className={`order-value order-status order-status--${o.Status}`}>
                    {o.Status}
                  </span>
                </div>
                <div className="order-row">
                  <span className="order-label">Date</span>
                  <span className="order-value">{formatDate(o.CreatedAt)}</span>
                </div>
              </div>
            ))}
          </div>
        )}

        <button className="button" onClick={() => setPage("home")}>
          Back to Home
        </button>
      </div>
    </div>
  );
}
