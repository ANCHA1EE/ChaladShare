import React, { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import axios from "axios";

import "../component/Login.css";
import bg from "../assets/bg.jpg";
import logo from "../assets/logo.png";

import { MdOutlineAlternateEmail, MdLockOutline } from "react-icons/md";
import { VscEye, VscEyeClosed } from "react-icons/vsc";
import { BiUser } from "react-icons/bi";

const Register = () => {
  const [formData, setForm] = useState({
    userEmail: "",
    username: "",
    password: "",
    confirmpassword: "",
  });

  const navigate = useNavigate();
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setshowConfirmPassword] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const validateEmail = (email) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);

  // ✅ เช็ค username ซ้ำ (ถ้า backend ยังไม่มี endpoint นี้ จะข้ามแบบไม่พัง)
  // const checkUsernameAvailable = async (username) => {
  //   try {
  //     const res = await axios.get("/auth/check-username", {
  //       params: { username },
  //       withCredentials: true,
  //       timeout: 8000,
  //     });
  //     // คาดหวังรูปแบบ { available: true/false }
  //     if (typeof res?.data?.available === "boolean") return res.data.available;
  //     // ถ้า backend ส่งแบบอื่น แต่ status 200 ให้ถือว่า "ผ่าน" (ไม่เดารูปแบบ)
  //     return true;
  //   } catch (err) {
  //     // ถ้าไม่มี route (404) หรือยังไม่ implement ให้ไม่บล็อก flow
  //     if (err?.response?.status === 404) return true;

  //     // ถ้า backend ตอบชัดว่า username ถูกใช้แล้ว
  //     const msg =
  //       err?.response?.data?.error ||
  //       err?.response?.data?.message ||
  //       err?.response?.data?.detail ||
  //       "";

  //     if (
  //       err?.response?.status === 409 ||
  //       /username/i.test(msg) ||
  //       /ชื่อผู้ใช้/i.test(msg) ||
  //       /ซ้ำ/i.test(msg) ||
  //       /ใช้แล้ว/i.test(msg)
  //     ) {
  //       return false;
  //     }

  //     // กรณีอื่น ๆ ไม่ให้พัง
  //     return true;
  //   }
  // };

  const handleChange = (e) => {
    setForm({
      ...formData,
      [e.target.name]: e.target.value,
    });

    // ✅ พอเริ่มแก้ไขให้ล้าง error (เฉพาะ UI)
    if (error) setError("");
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (loading) return;
    setError("");

    const email = formData.userEmail.trim().toLowerCase();
    const username = formData.username.trim();
    const password = formData.password;

    if (!email) {
      setError("กรุณากรอกอีเมล");
      return;
    }
    if (!validateEmail(email)) {
      setError("รูปแบบอีเมลไม่ถูกต้อง");
      return;
    }

    if (!username) {
      setError("กรุณากรอกชื่อผู้ใช้");
      return;
    }

    if (!password) {
      setError("กรุณากรอกรหัสผ่าน");
      return;
    }

    // รหัสผ่านต้องมากกว่า 8 ตัวอักษร (= อย่างน้อย 9)
    if (password.length < 8) {
      setError("รหัสผ่านต้องมากกว่า 8 ตัวอักษร");
      return;
    }

    if (password !== formData.confirmpassword) {
      setError("รหัสผ่านไม่ตรงกัน");
      return;
    }

    try {
      setLoading(true);

      // ✅ เช็ค username ซ้ำก่อน (ถ้ามี endpoint)
      // const ok = await checkUsernameAvailable(username);
      // if (!ok) {
      //   setError("ชื่อผู้ใช้นี้มีคนใช้แล้ว กรุณาเปลี่ยนชื่อผู้ใช้");
      //   return;
      // }

      await axios.post(
        "/auth/register/request-otp",
        { email, username }, // ✅ ส่ง username ไปด้วย
        {
          headers: { "Content-Type": "application/json" },
          withCredentials: true,
          timeout: 15000,
        }
      );

      navigate("/verify-otp", {
        state: { mode: "register", email, username, password },
        replace: true,
      });
    } catch (err) {
      const msg =
        err.response?.data?.error ||
        err.response?.data?.message ||
        err.response?.data?.detail ||
        "เกิดข้อผิดพลาดในการส่ง OTP";
      const status = err?.response?.status;

      if (status === 409) {
        // backend จะส่ง error เป็นข้อความไทยอยู่แล้ว เช่น "อีเมลนี้เคยสมัครไปแล้ว" / "ชื่อผู้ใช้นี้มีคนใช้แล้ว"
        if (/อีเมล/i.test(msg) || /email/i.test(msg)) {
          setError("อีเมลนี้เคยสมัครไปแล้ว");
          return;
        }
        if (/ชื่อผู้ใช้/i.test(msg) || /username/i.test(msg)) {
          setError("ชื่อผู้ใช้นี้มีคนใช้แล้ว กรุณาเปลี่ยนชื่อผู้ใช้");
          return;
        }
      }

      // ✅ map ข้อความให้ชัดขึ้นตาม requirement
      if (/email/i.test(msg) && /invalid|รูปแบบ/i.test(msg)) {
        setError("รูปแบบอีเมลไม่ถูกต้อง");
        return;
      }
      if (/username/i.test(msg) || /ชื่อผู้ใช้/i.test(msg)) {
        setError("ชื่อผู้ใช้นี้มีคนใช้แล้ว กรุณาเปลี่ยนชื่อผู้ใช้");
        return;
      }
      if (/password/i.test(msg) && (/8|length|ความยาว|มากกว่า/i.test(msg))) {
        setError("รหัสผ่านต้องมากกว่า 8 ตัวอักษร");
        return;
      }

      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="register-page">
      <div
        className="register-container"
        style={{
          backgroundImage: `url(${bg})`,
          backgroundSize: "cover",
          backgroundPosition: "center",
          backgroundRepeat: "no-repeat",
        }}
      >
        <div className="login-box">
          <img src={logo} alt="Logo" />
          <h2>สมัครสมาชิก</h2>

          {/* ✅ ปิด popup อังกฤษของ browser */}
          <form onSubmit={handleSubmit} noValidate>
            <div className="input-group">
              <span className="icon">
                <MdOutlineAlternateEmail />
              </span>
              <input
                type="text" // ✅ กัน tooltip ของ browser
                name="userEmail"
                value={formData.userEmail}
                onChange={handleChange}
                placeholder="Email"
                autoComplete="email"
                required
              />
            </div>

            <div className="input-group">
              <span className="icon">
                <BiUser />
              </span>
              <input
                type="text"
                name="username"
                value={formData.username}
                onChange={handleChange}
                placeholder="Username"
                autoComplete="username"
                required
              />
            </div>

            <div className="input-group" style={{ position: "relative" }}>
              <span className="icon">
                <MdLockOutline />
              </span>
              <input
                type={showPassword ? "text" : "password"}
                name="password"
                value={formData.password}
                onChange={handleChange}
                placeholder="Password"
                autoComplete="new-password"
                required
              />
              <span
                className="icon-right"
                onClick={() => setShowPassword(!showPassword)}
                style={{ cursor: "pointer" }}
              >
                {showPassword ? <VscEyeClosed /> : <VscEye />}
              </span>
            </div>

            <div className="input-group" style={{ position: "relative" }}>
              <span className="icon">
                <MdLockOutline />
              </span>
              <input
                type={showConfirmPassword ? "text" : "password"}
                name="confirmpassword"
                value={formData.confirmpassword}
                onChange={handleChange}
                placeholder="Confirm password"
                autoComplete="new-password"
                required
              />
              <span
                className="icon-right"
                onClick={() => setshowConfirmPassword(!showConfirmPassword)}
                style={{ cursor: "pointer" }}
              >
                {showConfirmPassword ? <VscEyeClosed /> : <VscEye />}
              </span>
            </div>

            {error && (
              <p style={{ color: "red", marginBottom: "10px" }}>{error}</p>
            )}

            <button
              type="submit"
              className="mb-3 p-2 border border-gray-300 rounded"
              disabled={loading}
            >
              {loading ? "กำลังส่ง OTP..." : "สมัครสมาชิก"}
            </button>
          </form>

          <div className="ClickToRegis">
            <p>คุณมีบัญชีแล้ว?</p>
            <Link to="/">เข้าสู่ระบบ</Link>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Register;