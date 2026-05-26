import { useState } from "react";

import "./../App.css";

export default function Checkout({
  setPage
}) {

  const [loading, setLoading] =
    useState(false);

  const [status, setStatus] =
    useState("");

  // ====================================
  // pay
  // ====================================

  const handlePay = async () => {

    setLoading(true);

    setStatus(
      "Processing payment..."
    );

    try {

      const res = await fetch(
        "http://localhost:8080/api/tickets/checkout",
        {
          method: "POST",

          headers: {
            "Content-Type":
              "application/json",
          },

          body: JSON.stringify({

            user_name: "john",

            phone_num: "0912345678",

            area: 1,

          }),
        }
      );

      const data = await res.json();

      console.log(data);

      // checkout success
      if (
        data.data?.checked_out
      ) {

        setPage("success");

      } else {

        setStatus(
          data.message ||
          "❌ Payment failed"
        );
      }

    } catch (err) {

      console.error(err);

      setStatus(
        "❌ Backend error"
      );

    } finally {

      setLoading(false);
    }
  };

  return (

    <div className="checkout-container">

      <div className="checkout-card">

        <h1>
          💳 Checkout
        </h1>

        <p>

          Your ticket has been reserved.

          <br />
          <br />

          Please complete payment
          within 5 minutes.

        </p>

        <input
          className="input"
          placeholder="Card Number"
        />

        <input
          className="input"
          placeholder="Card Holder"
        />

        <input
          className="input"
          placeholder="CVV"
        />

        <button
          className="button"

          disabled={loading}

          onClick={handlePay}
        >

          {loading
            ? "Processing..."
            : "Pay Now"}

        </button>

        <div className="status">

          {status}

        </div>

      </div>

    </div>
  );
}