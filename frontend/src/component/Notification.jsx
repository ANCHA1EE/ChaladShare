import React, { createContext, useCallback, useContext, useRef, useState } from "react";
import "./Notification.css";

const NotificationCtx = createContext(null);

function NotificationUI({ toast, leaving, onClose }) {
  if (!toast) return null;

  const cls =
    toast.variant === "success" ? "nt-success" :
    toast.variant === "error"   ? "nt-error" :
    toast.variant === "warning" ? "nt-warning" :
    "nt-info";

  return (
    <div className={`nt-wrap ${cls} ${leaving ? "nt-leave" : ""}`} role="status" aria-live="polite">
      <div className="nt-msg">{toast.message}</div>
      <button className="nt-x" onClick={onClose} aria-label="ปิด">✕</button>
    </div>
  );
}

export function NotificationProvider({ children }) {
  const [toast, setToast] = useState(null);
  const [leaving, setLeaving] = useState(false);

  const timerRef = useRef(null);
  const closeRef = useRef(null);

  const close = useCallback(() => {
    if (!toast) return;

    // เล่น slide-out ก่อน แล้วค่อยหาย
    setLeaving(true);

    if (timerRef.current) clearTimeout(timerRef.current);
    if (closeRef.current) clearTimeout(closeRef.current);

    closeRef.current = setTimeout(() => {
      setToast(null);
      setLeaving(false);
    }, 220); // ต้องตรงกับ nt-out 220ms ใน CSS
  }, [toast]);

  const notify = useCallback((variant, message, duration = 2500) => {
    // ถ้ามีอันเก่าอยู่ ให้เคลียร์ก่อน
    if (timerRef.current) clearTimeout(timerRef.current);
    if (closeRef.current) clearTimeout(closeRef.current);

    setLeaving(false);
    setToast({ variant, message });

    timerRef.current = setTimeout(() => {
      setLeaving(true);
      closeRef.current = setTimeout(() => {
        setToast(null);
        setLeaving(false);
      }, 220);
    }, duration);
  }, []);

  const success = useCallback((msg, ms) => notify("success", msg, ms), [notify]);
  const error   = useCallback((msg, ms) => notify("error", msg, ms), [notify]);
  const info    = useCallback((msg, ms) => notify("info", msg, ms), [notify]);
  const warning = useCallback((msg, ms) => notify("warning", msg, ms), [notify]);

  return (
    <NotificationCtx.Provider value={{ notify, success, error, info, warning, close }}>
      {children}
      <NotificationUI toast={toast} leaving={leaving} onClose={close} />
    </NotificationCtx.Provider>
  );
}

export function useNotification() {
  const ctx = useContext(NotificationCtx);
  if (!ctx) throw new Error("useNotification must be used within NotificationProvider");
  return ctx;
}