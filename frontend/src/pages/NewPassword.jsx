import React, { useMemo, useRef, useState } from "react";
import { useLocation, useNavigate, Link } from "react-router-dom";
import axios from "axios";
import "../component/AuthReset.css";
import bg from "../assets/bg.jpg";
import logo from "../assets/logo.png";

import { MdLockOutline } from "react-icons/md";
import { VscEye, VscEyeClosed, VscArrowLeft } from "react-icons/vsc";

const API_ORIGIN = process.env.REACT_APP_API_ORIGIN || "http://localhost:8080";

// ✅ แปลงข้อความ success/error จาก backend ให้เป็นไทย (แก้ที่หน้าบ้านอย่างเดียว)
const toThaiSuccess = (raw) => {
  const msg = String(raw || "").toLowerCase().trim();

  if (msg.includes("reset password success") || msg.includes("reset-password success")) {
    return "เปลี่ยนรหัสผ่านสำเร็จ";
  }
  if (msg.includes("success")) return "ดำเนินการสำเร็จ";

  // ถ้า backend ส่งไทยมาอยู่แล้ว ก็ใช้ได้
  return raw || "เปลี่ยนรหัสผ่านสำเร็จ";
};

const toThaiError = (raw) => {
  const msg = String(raw || "").toLowerCase().trim();

  if (msg.includes("invalid otp") || (msg.includes("otp") && msg.includes("expired"))) {
    return "รหัส OTP ไม่ถูกต้อง หรือหมดอายุแล้ว";
  }
  if (msg.includes("expired")) return "รหัส OTP หมดอายุแล้ว กรุณาส่งรหัสใหม่อีกครั้ง";
  if (msg.includes("invalid") && msg.includes("otp")) return "รหัส OTP ไม่ถูกต้อง";
  if (msg.includes("too many") || msg.includes("rate")) return "ทำรายการบ่อยเกินไป กรุณารอสักครู่แล้วลองใหม่";
  if (msg.includes("email")) return "อีเมลไม่ถูกต้อง หรือไม่พบผู้ใช้งาน";

  return raw || "เปลี่ยนรหัสผ่านไม่สำเร็จ";
};

export default function NewPassword() {
  const navigate = useNavigate();
  const location = useLocation();
  const timeoutRef = useRef(null);

  const { email, otp } = useMemo(() => {
    return {
      email: location.state?.email || "",
      otp: location.state?.otp || "",
    };
  }, [location.state]);

  const [pwd, setPwd] = useState("");
  const [confirm, setConfirm] = useState("");
  const [showPwd, setShowPwd] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const submit = async (e) => {
    e.preventDefault();
    setError("");
    setSuccess("");

    if (!email || !otp) {
      return setError("ข้อมูลไม่ครบ กรุณาเริ่มใหม่จากหน้า Forgot Password");
    }
    if (pwd.length < 8) {
      return setError("รหัสผ่านต้องมีอย่างน้อย 8 ตัวอักษร");
    }
    if (pwd !== confirm) {
      return setError("รหัสผ่านและยืนยันรหัสผ่านไม่ตรงกัน");
    }

    try {
      setLoading(true);

      const res = await axios.post(
        `${API_ORIGIN}/api/v1/auth/reset-password`,
        {
          email: String(email).trim().toLowerCase(),
          otp: String(otp),
          new_password: pwd,
        },
        { headers: { "Content-Type": "application/json" }, timeout: 15000 }
      );

      // ✅ บังคับให้เป็นไทย + ค้างไว้ 3 วินาที
      const rawMsg = res.data?.message || res.data?.success || "reset password success";
      setSuccess(toThaiSuccess(rawMsg));

      // (กันกดซ้ำ) ปิดปุ่มทันที
      // setLoading(true); // ไม่จำเป็น เพราะเรายัง setLoading อยู่แล้ว

      // ✅ ค้าง 3 วิ แล้วเด้งกลับหน้า login
      if (timeoutRef.current) clearTimeout(timeoutRef.current);
      timeoutRef.current = setTimeout(() => {
        navigate("/", { replace: true });
      }, 3000);
    } catch (err) {
      const raw =
        err.response?.data?.error ||
        err.response?.data?.message ||
        err.message ||
        "เปลี่ยนรหัสผ่านไม่สำเร็จ";
      setError(toThaiError(raw));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="reset-wrap" style={{ backgroundImage: `url(${bg})` }}>
      <div className="reset-card">
        <button
          className="reset-back"
          type="button"
          onClick={() => navigate(-1)}
          aria-label="ย้อนกลับ"
          disabled={loading}
        >
          <VscArrowLeft aria-hidden="true" />
        </button>

        <img className="reset-logo" src={logo} alt="Logo" />

        <h2 className="reset-title">ตั้งค่ารหัสผ่านใหม่</h2>

        <form onSubmit={submit}>
          <div className="pw-field">
            <span className="pw-lock" aria-hidden="true">
              <MdLockOutline />
            </span>

            <input
              className="pw-input"
              type={showPwd ? "text" : "password"}
              placeholder="รหัสผ่าน"
              value={pwd}
              onChange={(e) => setPwd(e.target.value)}
              disabled={loading}
              autoComplete="new-password"
            />

            <button
              type="button"
              className="pw-eye"
              onClick={() => setShowPwd((v) => !v)}
              aria-label="แสดง/ซ่อนรหัสผ่าน"
              disabled={loading}
            >
              {showPwd ? <VscEyeClosed /> : <VscEye />}
            </button>
          </div>

          <div className="pw-field">
            <span className="pw-lock" aria-hidden="true">
              <MdLockOutline />
            </span>

            <input
              className="pw-input"
              type={showConfirm ? "text" : "password"}
              placeholder="ยืนยันรหัสผ่าน"
              value={confirm}
              onChange={(e) => setConfirm(e.target.value)}
              disabled={loading}
              autoComplete="new-password"
            />

            <button
              type="button"
              className="pw-eye"
              onClick={() => setShowConfirm((v) => !v)}
              aria-label="แสดง/ซ่อนยืนยันรหัสผ่าน"
              disabled={loading}
            >
              {showConfirm ? <VscEyeClosed /> : <VscEye />}
            </button>
          </div>

          {error && <div className="reset-error">{error}</div>}
          {success && <div className="reset-success">{success}</div>}

          <button className="reset-primary" type="submit" disabled={loading}>
            {loading ? "กำลังบันทึก..." : "ยืนยัน"}
          </button>
        </form>

        <div className="reset-bottom">
          <Link to="/" className="reset-bottom-link">
            กลับไปเข้าสู่ระบบ
          </Link>
        </div>
      </div>
    </div>
  );
}