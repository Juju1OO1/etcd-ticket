import { useState } from "react";

import "./../App.css";

export default function Checkout({
  setPage
}) {

  const [loading, setLoading] =
    useState(false);

  const [status, setStatus] =
    useState("");


  const userName =
  localStorage.getItem("userName");

  const phoneNum =
  localStorage.getItem("phoneNum");

  const area =
  localStorage.getItem("selectedArea");

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

            user_name: userName,

            phone_num: phoneNum,

            area: Number(area),

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