import React, { useEffect, useState } from "react";
import { Navigate, Outlet, useLocation } from "react-router-dom";
import axios from "axios";

export default function RequireAuth() {
  const [ok, setOk] = useState(null); // null=กำลังเช็ค
  const location = useLocation();

  useEffect(() => {
    let alive = true;
    (async () => {
      try {
        await axios.get("/profile"); // protected -> จะ trigger refresh อัตโนมัติถ้า access หมด
        if (alive) setOk(true);
      } catch {
        if (alive) setOk(false);
      }
    })();
    return () => { alive = false; };
  }, []);

  if (ok === null) return <div>Loading...</div>;
  if (!ok) return <Navigate to="/" replace state={{ from: location }} />;
  return <Outlet />;
}