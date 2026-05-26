import "./../App.css";

export default function Success({setPage}) {

  return (

    <div className="success-container">

      <div className="success-card">

        <h1>
          ✅ Payment Successful
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
          onClick={() =>
            setPage("landing")
          }
        >
          Back to Home
        </button>

      </div>

    </div>
  );
}