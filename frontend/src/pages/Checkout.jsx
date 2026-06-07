import { useState, useEffect } from "react";

import "./../App.css";

export default function Checkout({
  setPage
}) {

  const [loading, setLoading] =
    useState(false);

  const [status, setStatus] =
    useState("");

  const [cardNumber, setCardNumber] =
    useState("");

  const [cardHolder, setCardHolder] =
    useState("");

  const [cvv, setCvv] =
    useState("");

  const [timeLeft, setTimeLeft] =
  useState(300);

  // 把 name 跟 phone 從 localstorage 拿出來
  const userName =
  localStorage.getItem("userName");

  const phoneNum =
  localStorage.getItem("phoneNum");

  const area =
  localStorage.getItem("selectedArea");

  // 倒計時
  useEffect(() => {
    let redirectTimeout;

    const timer = setInterval(() => {
      setTimeLeft((prev) => {
        if (prev <= 1) {
          clearInterval(timer);
          setStatus("⏰ Reservation expired, redirecting...");
          redirectTimeout = setTimeout(() => setPage("home"), 2000);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => {
      clearInterval(timer);
      clearTimeout(redirectTimeout);
    };
  }, []);

  const minutes =
    Math.floor(timeLeft / 60);

  const seconds =
    timeLeft % 60;

  const countdown =
    `${minutes}:${seconds
      .toString()
      .padStart(2, "0")}`;

  // ====================================
  // pay
  // ====================================

  const handlePay = async () => {

    if (!cardNumber.trim() || !cardHolder.trim() || !cvv.trim()) {
      setStatus("❌ Please fill in all card details");
      return;
    }

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
          within <strong>{countdown}</strong>.

        </p>

        <input
          className="input"
          placeholder="Card Number"
          value={cardNumber}
          onChange={(e) => setCardNumber(e.target.value)}
        />

        <input
          className="input"
          placeholder="Card Holder"
          value={cardHolder}
          onChange={(e) => setCardHolder(e.target.value)}
        />

        <input
          className="input"
          placeholder="CVV"
          value={cvv}
          onChange={(e) => setCvv(e.target.value)}
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