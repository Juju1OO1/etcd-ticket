import { useState } from "react";

import "./../App.css";

export default function UserInfo({
  setPage
}) {

  const [userName, setUserName] =
    useState("");

  const [phoneNum, setPhoneNum] =
    useState("");


  const handleNext = () => {

        localStorage.setItem(
        "userName",
        userName
      );

      localStorage.setItem(
        "phoneNum",
        phoneNum
      );

    setPage("home");
  };

  return (

    <div className="checkout-container">

      <div className="checkout-card">

        <h1>
          🎫 Your Information
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

        

        <button
          className="button"
          onClick={handleNext}
        >
          Register
        </button>

      </div>

    </div>
  );
}