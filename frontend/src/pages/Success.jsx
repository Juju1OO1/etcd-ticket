import "./../App.css";

export default function Success({setPage}) {

  return (

    <div className="success-container">

      <div className="success-card">

        <h1>
          🎉 Thank You 🎉
        </h1>

        <p>
          Your ticket purchase
          has been completed.
        </p>

        <div className="success-ticket">

          🎟️ Ticket Reserved

        </div>

        <button
          className="button"
          onClick={() => setPage("orders")}
        >
          View My Orders
        </button>

        <button
          className="button"
          style={{ background: "rgba(255,255,255,0.15)" }}
          onClick={() => setPage("home")}
        >
          Back to Home
        </button>

      </div>

    </div>
  );
}