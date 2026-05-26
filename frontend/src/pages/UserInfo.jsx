import { useState } from "react";

import "./../App.css";

export default function UserInfo({
  setPage
}) {

  const [userName, setUserName] =
    useState("");

  const [phoneNum, setPhoneNum] =
    useState("");

  const [area, setArea] =
    useState(1);

  const handleNext = () => {

    // 之後可以存進 context

    setPage("checkout");
  };

  return (

    <div className="checkout-container">

      <div className="checkout-card">

        <h1>
          🎫 Ticket Information
        </h1>

        <input
          className="input"

          placeholder="User Name"

          value={userName}

          onChange={(e) =>
            setUserName(e.target.value)
          }
        />

        <input
          className="input"

          placeholder="Phone Number"

          value={phoneNum}

          onChange={(e) =>
            setPhoneNum(e.target.value)
          }
        />

        <select
          className="input"

          value={area}

          onChange={(e) =>
            setArea(Number(e.target.value))
          }
        >

          <option value={1}>
            Area 1
          </option>

          <option value={2}>
            Area 2
          </option>

        </select>

        <button
          className="button"
          onClick={handleNext}
        >
          Continue to Payment
        </button>

      </div>

    </div>
  );
}