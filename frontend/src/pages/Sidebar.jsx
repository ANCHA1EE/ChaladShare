import React, { useLayoutEffect, useState } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { RiUser6Line, RiUserAddLine, RiLogoutCircleRLine, RiHome2Line } from "react-icons/ri";
import { HiOutlineSparkles } from "react-icons/hi2";
import { IoMdAddCircleOutline } from "react-icons/io";
import { TbHelpCircle } from "react-icons/tb";
import axios from "axios";

import "../component/Sidebar.css";
import logo from "../assets/logo.png";

const API_HOST = "http://localhost:8080/api/v1";

const Sidebar = () => {
  const navigate = useNavigate();
  const location = useLocation();

  const getKeyFromPath = (p) => {
    if (p === "/newpost") return "newpost";
    if (p === "/friends") return "friends";
    if (p === "/profile") return "profile";
    if (p === "/ai") return "ai";
    if (p === "/home") return "home";
    if (p === "/helper") return "helper";
    return "home";
  };

  // ✅ ตั้งค่าเริ่มต้นให้ตรงกับ path ตั้งแต่ render แรก
  const [activeKey, setActiveKey] = useState(() => getKeyFromPath(location.pathname));

  // ✅ ซิงค์ก่อน paint เพื่อลด/หายอาการกระพริบ
  useLayoutEffect(() => {
    setActiveKey(getKeyFromPath(location.pathname));
  }, [location.pathname]);

  const go = (key, path) => {
    setActiveKey(key);
    navigate(path);
  };

  const handleLogout = async () => {
    try {
      await axios.post(`${API_HOST}/auth/logout`, {}, { withCredentials: true });
    } catch (err) {
      console.log("logout error:", err);
    } finally {
      setActiveKey("home");
      navigate("/"); 
    }
  };

  return (
    <div className="sidebar">
      <div
        className="logo"
        onClick={() => go("home", "/home")}
        style={{ cursor: "pointer" }}
      >
        <img src={logo} alt="Chalad Share logo" />
      </div>

      <ul className="menu">
        <li className={activeKey === "home" ? "active" : ""} onClick={() => go("home", "/home")}>
          <RiHome2Line /> หน้าหลัก
        </li>

        <li className={activeKey === "newpost" ? "active" : ""} onClick={() => go("newpost", "/newpost")}>
          <IoMdAddCircleOutline /> สร้าง
        </li>

        <li className={activeKey === "ai" ? "active" : ""} onClick={() => go("ai", "/ai")}>
          <HiOutlineSparkles /> AI ช่วยสรุป
        </li>

        <li className={activeKey === "friends" ? "active" : ""} onClick={() => go("friends", "/friends")}>
          <RiUserAddLine /> เพื่อน
        </li>

        <li className={activeKey === "profile" ? "active" : ""} onClick={() => go("profile", "/profile")}>
          <RiUser6Line /> โปรไฟล์
        </li>

        <div className="menu-spacer" aria-hidden="true"></div>

        <li className={activeKey === "helper" ? "active" : ""} onClick={() => go("helper", "/helper")}>
          <TbHelpCircle /> คู่มือใช้งาน
        </li>

        <li onClick={handleLogout}>
          <RiLogoutCircleRLine /> ออกจากระบบ
        </li>
      </ul>
    </div>
  );
};

export default Sidebar;
