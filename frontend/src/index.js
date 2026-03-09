import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import reportWebVitals from './reportWebVitals';
import axios from "axios";
import './index.css';
import { NotificationProvider } from "./component/Notification";

const root = ReactDOM.createRoot(document.getElementById('root'));
axios.defaults.baseURL =
  process.env.REACT_APP_API_URL || "http://localhost:8080/api/v1";
axios.defaults.withCredentials = true;

let refreshPromise = null;

axios.interceptors.response.use(
  (r) => r,
  async (err) => {
    const status = err?.response?.status;
    const original = err?.config || {};
    const url = original?.url || ""; // เช่น "/auth/login"

    // ไม่ redirect สำหรับ auth endpoints (ให้หน้าแสดง error เอง)
    const skipAuth =
      url.includes("/auth/login") ||
      url.includes("/auth/register") ||
      url.includes("/auth/logout") ||
      url.includes("/auth/refresh") ||         //*    
      // url.includes("/auth/register/request-otp") ||  //*
      // url.includes("/auth/register/confirm-otp") || //*
      url.includes("/auth/verify-otp") ||
      url.includes("/auth/forgot-password") ||
      url.includes("/auth/reset-password");

    if (status !== 401 || skipAuth) {
      return Promise.reject(err);
    }

    if (original._retry) {
      window.location.replace("/");
      return Promise.reject(err);
    }
    original._retry = true;

    try {
      // ทำให้ refresh เกิดได้ทีละ 1 ครั้ง
      if (!refreshPromise) {
        refreshPromise = axios
          .post("/auth/refresh", {}, { withCredentials: true })
          .finally(() => {
            refreshPromise = null;
          });
      }

      await refreshPromise;

      // retry request เดิม
      return axios(original);
    } catch (refreshErr) {
      window.location.replace("/");
      return Promise.reject(refreshErr);
    }
  }
);

root.render(
  <React.StrictMode>
    {/* ✅ ครอบ App ด้วย Provider */}
    <NotificationProvider>
      <App />
    </NotificationProvider>
  </React.StrictMode>
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
