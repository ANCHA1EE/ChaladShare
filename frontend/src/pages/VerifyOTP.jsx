import React, { useEffect, useMemo, useRef, useState } from "react";
import { useLocation, useNavigate, Link } from "react-router-dom";
import { VscArrowLeft } from "react-icons/vsc";
import { useNotification } from "../component/Notification";
import axios from "axios";
import bg from "../assets/bg.jpg";
import "../component/AuthReset.css";

const API_ORIGIN = process.env.REACT_APP_API_ORIGIN || "http://localhost:8080";
const OTP_TTL_SECONDS = 180;

const pad2 = (n) => String(n).padStart(2, "0");
const formatMMSS = (sec) => `${pad2(Math.floor(sec / 60))}:${pad2(sec % 60)}`;

// ✅ แปล error จาก backend เป็นภาษาไทย (แก้ที่หน้าบ้านอย่างเดียว)
const toThaiError = (raw) => {
  const msg = String(raw || "").toLowerCase().trim();

  // เคสหลักที่เจอ
  if (msg.includes("invalid otp") || (msg.includes("otp") && msg.includes("expired"))) {
    return "รหัส OTP ไม่ถูกต้อง หรือหมดอายุแล้ว";
  }

  // เผื่อ backend ส่งรูปแบบอื่น
  if (msg.includes("expired")) return "รหัส OTP หมดอายุแล้ว กรุณาส่งรหัสใหม่อีกครั้ง";
  if (msg.includes("invalid") && msg.includes("otp")) return "รหัส OTP ไม่ถูกต้อง";
  if (msg.includes("too many") || msg.includes("rate")) return "ส่งรหัสบ่อยเกินไป กรุณารอสักครู่แล้วลองใหม่";
  if (msg.includes("email")) return "อีเมลไม่ถูกต้อง หรือไม่พบผู้ใช้งาน";

  // ไม่ให้หลุดอังกฤษ (เลือกอย่างใดอย่างหนึ่ง)
  // return raw || "เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง";
  return "เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง";
};

export default function VerifyOTP() {
  const navigate = useNavigate();
  const location = useLocation();
  const { success: notifySuccess, error: notifyError } = useNotification();

  // ✅ mode: "forgot" | "register"
  const mode = useMemo(() => {
    const m = location.state?.mode;
    return m === "register" ? "register" : "forgot";
  }, [location.state]);

  const initialEmail = useMemo(() => {
    const e = location.state?.email;
    return typeof e === "string" ? e : "";
  }, [location.state]);

  const navTimerRef = useRef(null);

  useEffect(() => {
    return () => {
      if (navTimerRef.current) clearTimeout(navTimerRef.current);
    };
  }, []);

  const registerPayload = useMemo(() => {
    return {
      username: typeof location.state?.username === "string" ? location.state.username : "",
      password: typeof location.state?.password === "string" ? location.state.password : "",
    };
  }, [location.state]);

  const [email] = useState(initialEmail);
  const [digits, setDigits] = useState(["", "", "", "", "", ""]);
  const [remaining, setRemaining] = useState(OTP_TTL_SECONDS);

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const inputRefs = useRef([]);
  const timerRef = useRef(null);

  useEffect(() => {
    if (!initialEmail) {
      navigate(mode === "register" ? "/register" : "/forgot_password", { replace: true });
      return;
    }
    if (mode === "register" && (!registerPayload.username || !registerPayload.password)) {
      navigate("/register", { replace: true });
    }
  }, [initialEmail, mode, registerPayload.username, registerPayload.password, navigate]);

  useEffect(() => {
    setRemaining(OTP_TTL_SECONDS);
    if (timerRef.current) clearInterval(timerRef.current);
    timerRef.current = setInterval(() => {
      setRemaining((prev) => (prev > 0 ? prev - 1 : 0));
    }, 1000);
    return () => timerRef.current && clearInterval(timerRef.current);
  }, []);

  const focusIndex = (i) => {
    const el = inputRefs.current[i];
    if (el) el.focus();
  };

  const onChangeAt = (i, val) => {
    setError("");
    const v = val.replace(/\D/g, "").slice(-1);
    const next = [...digits];
    next[i] = v;
    setDigits(next);
    if (v && i < 5) focusIndex(i + 1);
  };

  // ✅ Backspace แบบไม่ preventDefault
  const onKeyDownAt = (i, e) => {
    if (e.key !== "Backspace") return;
    setError("");

    const next = [...digits];

    if (next[i]) {
      next[i] = "";
      setDigits(next);
      return;
    }

    if (i > 0) {
      next[i - 1] = "";
      setDigits(next);
      focusIndex(i - 1);
    }
  };

  const onPaste = (e) => {
    const text = e.clipboardData.getData("text").replace(/\D/g, "").slice(0, 6);
    if (!text) return;
    e.preventDefault();

    const next = ["", "", "", "", "", ""];
    for (let i = 0; i < text.length; i++) next[i] = text[i];
    setDigits(next);
    setError("");
    focusIndex(Math.min(text.length, 5));
  };

  const resendOtp = async () => {
    if (!email || loading) return;
    setError("");

    try {
      setLoading(true);

      const body = { email: email.trim().toLowerCase() };
      const url =
        mode === "register"
          ? `${API_ORIGIN}/api/v1/auth/register/request-otp`
          : `${API_ORIGIN}/api/v1/auth/forgot-password`;

      await axios.post(url, body, {
        headers: { "Content-Type": "application/json" },
        timeout: 15000,
      });

      setDigits(["", "", "", "", "", ""]);
      setRemaining(OTP_TTL_SECONDS);

      if (timerRef.current) clearInterval(timerRef.current);
      timerRef.current = setInterval(() => {
        setRemaining((prev) => (prev > 0 ? prev - 1 : 0));
      }, 1000);

      focusIndex(0);
    } catch (err) {
      const raw = err.response?.data?.error || err.response?.data?.message || err.message || "ส่งรหัสใหม่ไม่สำเร็จ";
      setError(toThaiError(raw));
    } finally {
      setLoading(false);
    }
  };

  const otpNow = digits.join("");

  const verifyOtp = async () => {
    setError("");

    if (remaining <= 0) return setError("รหัส OTP หมดอายุแล้ว กรุณาส่งรหัสใหม่อีกครั้ง");
    if (!/^\d{6}$/.test(otpNow)) return setError("กรุณากรอกรหัส OTP ให้ครบ 6 หลัก");

    try {
      setLoading(true);

      if (mode === "forgot") {
        await axios.post(
          `${API_ORIGIN}/api/v1/auth/forgot-password/verify-otp`,
          { email: email.trim().toLowerCase(), otp: otpNow },
          { headers: { "Content-Type": "application/json" }, timeout: 15000 }
        );

        navigate("/new-password", {
          state: { email, otp: otpNow, ttlLeft: remaining },
          replace: true,
        });
        return;
      }

      const confirmRes = await axios.post(
        `${API_ORIGIN}/api/v1/auth/register/confirm-otp`,
        { email: email.trim().toLowerCase(), otp: otpNow },
        { headers: { "Content-Type": "application/json" }, timeout: 15000 }
      );

      const verify_token = confirmRes.data?.verify_token;
      if (!verify_token) {
        setError("ไม่สามารถยืนยัน OTP ได้ (verify_token หาย)");
        return;
      }

      await axios.post(
        `${API_ORIGIN}/api/v1/auth/register`,
        {
          email: email.trim().toLowerCase(),
          username: registerPayload.username,
          password: registerPayload.password,
          verify_token,
        },
        { headers: { "Content-Type": "application/json" }, timeout: 15000 }
      );

      notifySuccess("สมัครสมาชิกสำเร็จ", 3000);

      if (navTimerRef.current) clearTimeout(navTimerRef.current);
      navTimerRef.current = setTimeout(() => {
        navigate("/home", { replace: true });
      }, 800);
    } catch (err) {
      if (navTimerRef.current) clearTimeout(navTimerRef.current);

      const raw = err.response?.data?.error || err.response?.data?.message || err.message;
      const msg = toThaiError(raw);
      
      setError(toThaiError(raw));
      setError(msg);       // ข้อความแดงในหน้า
      notifyError(msg, 2500); // toast มุมขวาบน
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

        <h2 className="reset-title">ยืนยันตัวตน</h2>

        <div className="otp-row" onPaste={onPaste}>
          {digits.map((d, i) => (
            <input
              key={i}
              ref={(el) => (inputRefs.current[i] = el)}
              className="otp-box"
              value={d}
              onChange={(e) => onChangeAt(i, e.target.value)}
              onKeyDown={(e) => onKeyDownAt(i, e)}
              onFocus={() => setError("")}
              inputMode="numeric"
              maxLength={1}
              disabled={loading}
              aria-label={`OTP หลักที่ ${i + 1}`}
            />
          ))}
        </div>

        <p className="reset-sub">โปรดตรวจสอบอีเมลของคุณ กรุณาระบุรหัส OTP ที่ได้รับภายใน 3 นาที</p>

        <div className="reset-timer">
          หมดอายุใน <b>{formatMMSS(remaining)}</b>
        </div>

        {error && <div className="reset-error">{error}</div>}

        <button className="reset-primary" type="button" onClick={verifyOtp} disabled={loading}>
          {loading ? "กำลังตรวจสอบ..." : "ยืนยัน"}
        </button>

        <div className="reset-footer">
          <span>ยังไม่ได้รับ OTP ใช่ไหม ?</span>
          <button className="reset-link" type="button" onClick={resendOtp} disabled={loading}>
            ส่งรหัสใหม่อีกครั้ง
          </button>
        </div>

        <div className="reset-bottom">
          <Link to="/" className="reset-bottom-link">
            กลับไปเข้าสู่ระบบ
          </Link>
        </div>
      </div>
    </div>
  );
}