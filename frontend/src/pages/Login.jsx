import React, { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import axios from "axios";

import { MdOutlineAlternateEmail, MdLockOutline } from "react-icons/md";
import { VscEye, VscEyeClosed } from "react-icons/vsc";

import "../component/Login.css";
import bg from "../assets/bg.jpg";
import logo from "../assets/logo.png";

const Login = () => {
  const [formData, setFormData] = useState({ email: "", password: "" });
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const validateEmail = (email) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);

  const handleChange = (e) => {
    const { name, value } = e.target;

    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));

    // ล้าง error เมื่อผู้ใช้เริ่มแก้
    if (error) setError("");
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (loading) return;

    setError("");

    const email = formData.email.trim().toLowerCase();
    const password = formData.password;

    // 🔹 Validate ตอน submit เท่านั้น
    if (!email) {
      setError("กรุณากรอกอีเมล");
      return;
    }

    if (!validateEmail(email)) {
      setError("รูปแบบอีเมลไม่ถูกต้อง");
      return;
    }

    if (!password) {
      setError("กรุณากรอกรหัสผ่าน");
      return;
    }

    try {
      setLoading(true);

      const res = await axios.post(
        "/auth/login",
        { email, password },
        {
          headers: { "Content-Type": "application/json" },
          withCredentials: true, //ให้รับ/ส่ง cookie
          timeout: 15000,
        },
      );

      if (res.status === 200) {
        navigate("/home");
      }
    } catch (err) {
      const status = err?.response?.status;

      if (status === 401) {
        setError("อีเมลหรือรหัสผ่านไม่ถูกต้อง");
        return;
      }

      const raw =
        err?.response?.data?.error ||
        err?.response?.data?.detail ||
        err?.message ||
        "เข้าสู่ระบบไม่สำเร็จ";

      setError(raw);
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
          <img src={logo} alt="Logo" />
          <h2>เข้าสู่ระบบ</h2>

          {/* ปิด browser validation */}
          <form onSubmit={handleSubmit} noValidate>
            <div className="input-group">
              <span className="icon">
                <MdOutlineAlternateEmail />
              </span>
              <input
                type="text" // เปลี่ยนจาก email เพื่อกัน popup อังกฤษ
                name="email"
                value={formData.email}
                onChange={handleChange}
                placeholder="Email"
                autoComplete="email"
              />
            </div>

            <div className="input-group">
              <span className="icon">
                <MdLockOutline />
              </span>
              <input
                type={showPassword ? "text" : "password"}
                name="password"
                value={formData.password}
                onChange={handleChange}
                placeholder="Password"
                autoComplete="current-password"
              />
              <span
                className="icon-right"
                onClick={() => setShowPassword((s) => !s)}
                role="button"
                tabIndex={0}
              >
                {showPassword ? <VscEyeClosed /> : <VscEye />}
              </span>
            </div>

            <div className="forgot-password">
              <Link to="/forgot_password">ลืมรหัสผ่าน?</Link>
            </div>

            <button type="submit" disabled={loading}>
              {loading ? "กำลังเข้าสู่ระบบ..." : "เข้าสู่ระบบ"}
            </button>

            {error && (
              <p style={{ color: "red", fontSize: "15px", marginTop: "10px" }}>
                {error}
              </p>
            )}
          </form>

          <div className="ClickToRegis">
            <p>มีบัญชีแล้วหรือยัง?</p>
            <Link to="/register">สมัครสมาชิก</Link>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Login;
