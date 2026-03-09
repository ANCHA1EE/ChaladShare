import React, { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import axios from "axios";

import { MdOutlineAlternateEmail } from "react-icons/md";
import { VscArrowLeft } from "react-icons/vsc";

import "../component/Login.css"; // ใช้ CSS เดียวกับหน้า Login
import bg from "../assets/bg.jpg";
import logo from "../assets/logo.png";

// ✅ ตั้งค่า base URL ของ backend (dev)
// ถ้าตั้ง .env ฝั่ง frontend ไว้ก็ใช้ REACT_APP_API_ORIGIN ได้ เช่น http://localhost:8080
const API_ORIGIN = process.env.REACT_APP_API_ORIGIN || "http://localhost:8080";

const ForgotPassword = () => {
  const navigate = useNavigate(); // ✅ เพิ่ม (เกี่ยวกับการไปหน้า reset)
  const [email, setEmail] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [loading, setLoading] = useState(false);

  const validateEmail = (val) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(val);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setSuccess("");

    const normalized = email.trim().toLowerCase();
    if (!normalized) return setError("กรุณากรอกอีเมล");
    if (!validateEmail(normalized)) return setError("รูปแบบอีเมลไม่ถูกต้อง");

    try {
      setLoading(true);

      // ✅ ยิงไป backend จริง
      const url = `${API_ORIGIN}/api/v1/auth/forgot-password`;
      const res = await axios.post(
        url,
        { email: normalized },
        {
          headers: { "Content-Type": "application/json" },
          // withCredentials: true, // ถ้า backend ใช้ cookie ค่อยเปิด
          timeout: 15000,
        }
      );

      // backend ของบีมตอบ 200 พร้อม message เสมอ
      const msg =
        res.data?.message ||
        "ส่งคำขอแล้ว หากมีอีเมลนี้ในระบบ คุณจะได้รับรหัส OTP ";

      setSuccess(msg);

      // ✅ ไปหน้า ResetPassword พร้อมส่ง email ไปด้วย (ไม่แก้ส่วนอื่น)
      setTimeout(() => {
        navigate("/verify-otp", { state: { mode: "forgot", email: normalized }, replace: true });
      }, 3000);
    } catch (err) {
      // ✅ debug ให้รู้สาเหตุจริง (404 / CORS / 500)
      console.log("FORGOT_PASSWORD_ERROR:", err);
      console.log("STATUS:", err.response?.status);
      console.log("DATA:", err.response?.data);

      // ถ้าเป็น CORS บางที err.response จะไม่มี
      if (!err.response) {
        setError(
          "เชื่อมต่อเซิร์ฟเวอร์ไม่ได้ (อาจเป็น CORS หรือ backend ไม่ได้รันอยู่) กรุณาลองใหม่"
        );
      } else {
        // ✅ เพิ่ม: ถ้าอีเมลไม่มีในระบบ ให้แจ้ง user ตรงๆ และห้ามไปหน้า OTP
        if (err.response.status === 404) {
          setError("ไม่มีอีเมลนี้ในระบบ กรุณาตรวจสอบอีเมลอีกครั้ง");
          return;
        }

        const msg =
          err.response?.data?.error ||
          err.response?.data?.message ||
          err.response?.data?.detail ||
          "ไม่สามารถส่งรหัส OTP ได้";
        setError(msg);
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-page">
      <div
        className="login-container"
        style={{
          backgroundImage: `url(${bg})`,
          backgroundSize: "cover",
          backgroundPosition: "center",
          backgroundRepeat: "no-repeat",
        }}
      >
        <div className="login-box">
          {/* ปุ่มย้อนกลับ */}
          <button
            className="back-circle-btn"
            onClick={() => navigate(-1)}
            aria-label="ย้อนกลับ"
            type="button"
          >
            <VscArrowLeft aria-hidden="true" />
          </button>


          <img src={logo} alt="Logo" />
          <h2>ลืมรหัสผ่าน</h2>

          <form onSubmit={handleSubmit} noValidate>
            <div className="input-group">
              <span className="icon">
                <MdOutlineAlternateEmail />
              </span>
              <input
                type="email"
                name="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="อีเมล"
                autoComplete="email"
                required
                disabled={loading}
              />
            </div>
            {/* โปรดตรวจสอบอีเมลของท่าน <br />
          กรุณาระบุรหัส OTP ที่ได้รับภายใน <b>3 นาที</b> */}
            <p style={{ color: "#6b7280", fontSize: 13, marginTop: -4 }}>
              ในขั้นตอนถัดไป รหัส OTP จะถูกส่งไปทางอีเมลของคุณ <br />
              โปรดตรวจสอบอีเมล <br />
            </p>

            <button type="submit" disabled={loading}>
              {loading ? "กำลังส่งรหัส OTP..." : "ส่ง"}
            </button>

            {error && (
              <p style={{ color: "red", fontSize: 15, marginTop: "0.75rem" }}>
                {error}
              </p>
            )}

            {success && (
              <p
                style={{
                  color: "#0f5132",
                  background: "#d1e7dd",
                  border: "1px solid #badbcc",
                  padding: "8px 10px",
                  borderRadius: 8,
                  fontSize: 14,
                  marginTop: "0.75rem",
                }}
              >
                {success}
              </p>
            )}
          </form>

          <div className="ClickToRegis" style={{ marginTop: 12 }}>
            <p>จำรหัสผ่านได้แล้ว?</p>
            <Link to="/">กลับไปเข้าสู่ระบบ</Link>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ForgotPassword;
