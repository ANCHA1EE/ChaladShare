// หน้า .jsx (ทำ prefix แล้ว)

import React, { useEffect, useState, useCallback } from "react";
import { IoSearch } from "react-icons/io5";
import { useNavigate } from "react-router-dom";
import { FaArrowLeft } from "react-icons/fa";
import Sidebar from "./Sidebar";
import axios from "axios";
import Footer from "../component/Footer";
import picdefault from "../assets/default.png";
import "../component/Friends.css";

// const API_HOST = "http://localhost:8080";
const API_URL =
  process.env.REACT_APP_API_URL || "http://localhost:8080/api/v1";

const FILE_BASE_URL = API_URL.replace(/\/api\/v1\/?$/, "");

const toAbsUrl = (p) => {
  if (!p) return "";
  if (p.startsWith("http")) return p;
  const clean = p.replace(/^\.\//, "").replace(/^\./, "");
  return `${FILE_BASE_URL}${clean.startsWith("/") ? clean : `/${clean}`}`;
};

const Friends = () => {
  const [ownerId, setOwnerId] = useState(null);
  const [activeTab, setActiveTab] = useState("my");
  const [query, setQuery] = useState("");
  const [page, setPage] = useState(1);
  const size = 20;

  const [friends, setFriends] = useState([]);
  const [totalFriends, setTotalFriends] = useState(0);
  const [loadingFriends, setLoadingFriends] = useState(false);

  const [searchUsers, setSearchUsers] = useState([]);
  const [searchTotal, setSearchTotal] = useState(0);
  const [loadingSearch, setLoadingSearch] = useState(false);

  const [incoming, setIncoming] = useState([]);
  const [loadingReq, setLoadingReq] = useState(false);

  const [debouncedQuery, setDebouncedQuery] = useState("");
  const isSearching = activeTab === "my" && query.trim().length > 0;

  const navigate = useNavigate();

  useEffect(() => {
    const fetchMe = async () => {
      try {
        const { data } = await axios.get("/profile");
        const id = data.user_id || data.id;
        if (id) setOwnerId(id);
      } catch (err) {
        console.error("Fetch profile failed", err);
      }
    };
    fetchMe();
  }, []);

  const fetchFriends = useCallback(
    async (q, p) => {
      if (!ownerId) return;
      setLoadingFriends(true);
      try {
        const { data } = await axios.get(`/social/friends/${ownerId}`, {
          params: { search: q, page: p, size },
        });
        setFriends(data.items || []);
        setTotalFriends(data.total || 0);
      } catch (e) {
        console.error("listFriends:", e);
        alert("โหลดรายชื่อเพื่อนไม่สำเร็จ");
      } finally {
        setLoadingFriends(false);
      }
    },
    [ownerId, size],
  );

  const fetchUserSearch = useCallback(
    async (q, p) => {
      const qq = (q || "").trim();

      if (qq.length === 0) {
        setSearchUsers([]);
        setSearchTotal(0);
        return;
      }

      setLoadingSearch(true);
      try {
        const { data } = await axios.get(`/social/addfriends`, {
          params: { search: qq, page: p, size },
        });
        setSearchUsers(data.items || []);
        setSearchTotal(data.total || 0);
      } catch (e) {
        console.error("userSearch:", e);
        alert("ค้นหาเพื่อนไม่สำเร็จ");
      } finally {
        setLoadingSearch(false);
      }
    },
    [size],
  );

  const fetchIncoming = useCallback(async () => {
    setLoadingReq(true);
    try {
      const { data } = await axios.get(`/social/requests/incoming`, {
        params: { page: 1, size: 50 },
      });
      setIncoming(data.items || []);
    } catch (e) {
      console.error("incoming:", e);
      alert("โหลดคำขอเป็นเพื่อนไม่สำเร็จ");
    } finally {
      setLoadingReq(false);
    }
  }, []);

  const unfriend = async (targetId) => {
    try {
      await axios.delete(`/social/friends/${targetId}`);
      setFriends((prev) => prev.filter((it) => it.user_id !== targetId));
      setTotalFriends((t) => Math.max(0, t - 1));

      if (isSearching) fetchUserSearch(query, page);
    } catch (e) {
      console.error("unfriend:", e);
      alert("ลบเพื่อนไม่สำเร็จ");
    }
  };

  const sendRequest = async (targetId) => {
    try {
      await axios.post(`/social/requests`, { to_user_id: targetId });
      fetchUserSearch(query, page);
      fetchIncoming();
    } catch (e) {
      console.error("sendRequest:", e);
      alert("ส่งคำขอเป็นเพื่อนไม่สำเร็จ");
    }
  };

  // requests tab
  const acceptRequest = async (requestId) => {
    try {
      await axios.post(`/social/requests/${requestId}/accept`);
      await fetchIncoming();
      fetchFriends("", 1);
      if (isSearching) fetchUserSearch(query, page);
    } catch (e) {
      console.error("accept:", e);
      alert("ยอมรับคำขอไม่สำเร็จ");
    }
  };

  const declineRequest = async (requestId) => {
    try {
      await axios.post(`/social/requests/${requestId}/decline`);
      await fetchIncoming();
      if (isSearching) fetchUserSearch(query, page);
    } catch (e) {
      console.error("decline:", e);
      alert("ปฏิเสธคำขอไม่สำเร็จ");
    }
  };

  // search
  const acceptFromSearch = async (requestId) => {
    try {
      await axios.post(`/social/requests/${requestId}/accept`);
      await fetchIncoming();
      fetchFriends("", 1);
      fetchUserSearch(query, page);
    } catch (e) {
      console.error("acceptFromSearch:", e);
      alert("ยอมรับคำขอไม่สำเร็จ");
    }
  };

  const declineFromSearch = async (requestId) => {
    try {
      await axios.post(`/social/requests/${requestId}/decline`);
      await fetchIncoming();
      fetchUserSearch(query, page);
    } catch (e) {
      console.error("declineFromSearch:", e);
      alert("ปฏิเสธคำขอไม่สำเร็จ");
    }
  };

  const cancelOutgoingFromSearch = async (requestId) => {
    try {
      await axios.delete(`/social/requests/${requestId}`);
      await fetchIncoming();
      fetchUserSearch(query, page);
    } catch (e) {
      console.error("cancelOutgoingFromSearch:", e);
      alert("ยกเลิกคำขอไม่สำเร็จ");
    }
  };

  const goToProfile = (userId) => {
    if (userId) {
      navigate(`/profile/${userId}`);
    }
  };

  const items = isSearching ? searchUsers : friends;
  const total = isSearching ? searchTotal : totalFriends;
  const loadingList = isSearching ? loadingSearch : loadingFriends;
  const totalPages = Math.max(1, Math.ceil((total || 0) / size));

  useEffect(() => {
    const t = setTimeout(() => setDebouncedQuery(query.trim()), 300);
    return () => clearTimeout(t);
  }, [query]);

  useEffect(() => {
    if (activeTab !== "my") return;
    setPage(1);
  }, [activeTab, debouncedQuery]);

  useEffect(() => {
    if (!ownerId) return;
    if (activeTab !== "my") return;

    if (debouncedQuery.length === 0) {
      fetchFriends("", page);
    } else {
      fetchUserSearch(debouncedQuery, page);
    }
  }, [ownerId, activeTab, debouncedQuery, page, fetchFriends, fetchUserSearch]);

  useEffect(() => {
    if (!ownerId) return;
    fetchIncoming();
  }, [ownerId, fetchIncoming]);

  useEffect(() => {
    if (!ownerId) return;
    if (activeTab === "requests") fetchIncoming();
  }, [ownerId, activeTab, fetchIncoming]);

  return (
    <div className="friends-page">
      <div className="friends-container">
        <Sidebar />

        <main className="friends-main">
          {/*Top bar= */}
          <div className="friends-topbar">
            <div className="friends-top-left">
              {activeTab === "requests" && (
                <button
                  type="button"
                  className="back-btn"
                  onClick={() => setActiveTab("my")}
                  aria-label="ย้อนกลับ"
                >
                  <FaArrowLeft />
                </button>
              )}

              <h2 className="friends-title">
                {activeTab === "my" && "เพื่อนของฉัน"}
                {activeTab === "requests" && "คำขอเป็นเพื่อน"}
              </h2>

              <div className="friends-actions">
                <button
                  type="button"
                  className={`friends-pill friends-pill--outline ${
                    activeTab === "requests" ? "is-active" : ""
                  }`}
                  onClick={() => setActiveTab("requests")}
                >
                  คำขอ ({incoming.length})
                </button>
              </div>
            </div>

            {activeTab === "my" && (
              <div className="friends-search">
                <input
                  type="text"
                  placeholder="ค้นหาเพื่อน"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                />
                <IoSearch className="friends-search-icon" />
              </div>
            )}
          </div>

          {/* TAB: my */}
          {activeTab === "my" && (
            <>
              {loadingList && (
                <div className="friends-placeholder">
                  {isSearching ? "กำลังค้นหา..." : "กำลังโหลด..."}
                </div>
              )}

              {!loadingList && (
                <>
                  <ul className="friends-list">
                    {items.map((u) => (
                      <li key={u.user_id} className="friends-item">
                        <div
                          className="friends-left"
                          onClick={() => goToProfile(u.user_id)}
                          style={{ cursor: "pointer" }}
                          title="ไปที่โปรไฟล์"
                        >
                          <img
                            className="friends-avatar"
                            src={toAbsUrl(u.avatar) || picdefault}
                            alt={`${u.username || u.user_id} avatar`}
                            onError={(e) => (e.currentTarget.src = picdefault)}
                          />
                          <div className="friends-name">
                            <span className="friends-name-main">
                              {u.username || `user#${u.user_id}`}
                            </span>
                          </div>
                        </div>

                        {/* Right action button*/}
                        {!isSearching ? (
                          // โหมดเพื่อนของฉัน: ลบเพื่อน
                          <button
                            className="friends-remove"
                            onClick={() => unfriend(u.user_id)}
                          >
                            ลบเพื่อน
                          </button>
                        ) : (
                          // โหมดค้นหาทั้งระบบ: แสดงปุ่มตามสถานะ
                          <>
                            {u.is_friend ? (
                              <button
                                className="friends-pill friends-pill--disabled"
                                disabled
                              >
                                เป็นเพื่อนแล้ว
                              </button>
                            ) : u.request_status === "pending" &&
                              u.request_direction === "incoming" ? (
                              <div className="friends-actions-right">
                                <button
                                  className="friends-pill friends-pill--green"
                                  onClick={() => acceptFromSearch(u.request_id)}
                                >
                                  ยอมรับ
                                </button>
                                <button
                                  className="friends-pill friends-pill--danger"
                                  onClick={() =>
                                    declineFromSearch(u.request_id)
                                  }
                                >
                                  ปฏิเสธ
                                </button>
                              </div>
                            ) : u.request_status === "pending" &&
                              u.request_direction === "outgoing" ? (
                              <div className="friends-actions-right">
                                <button
                                  className="friends-pill friends-pill--disabled"
                                  disabled
                                >
                                  รอการตอบรับ
                                </button>
                                <button
                                  className="friends-remove"
                                  onClick={() =>
                                    cancelOutgoingFromSearch(u.request_id)
                                  }
                                >
                                  ยกเลิก
                                </button>
                              </div>
                            ) : (
                              <button
                                className="friends-pill friends-pill--primary"
                                onClick={() => sendRequest(u.user_id)}
                              >
                                เพิ่มเพื่อน
                              </button>
                            )}
                          </>
                        )}
                      </li>
                    ))}

                    {items.length === 0 && (
                      <div className="friends-placeholder">
                        {isSearching
                          ? "ไม่พบผู้ใช้ที่ตรงกับคำค้น"
                          : "ยังไม่มีเพื่อน"}
                      </div>
                    )}
                  </ul>

                  {Number.isFinite(total) && totalPages > 1 && (
                    <div className="friends-pagination">
                      <button
                        disabled={page <= 1}
                        onClick={() => setPage((p) => p - 1)}
                      >
                        ก่อนหน้า
                      </button>
                      <span>
                        {page} / {totalPages}
                      </span>
                      <button
                        disabled={page >= totalPages}
                        onClick={() => setPage((p) => p + 1)}
                      >
                        ถัดไป
                      </button>
                    </div>
                  )}
                </>
              )}
            </>
          )}

          {/* TAB requests */}
          {activeTab === "requests" && (
            <>
              {loadingReq && (
                <div className="friends-placeholder">กำลังโหลดคำขอ...</div>
              )}

              {!loadingReq && (
                <ul className="friends-list">
                  {incoming.map((r) => (
                    <li key={r.request_id} className="friends-item">
                      {/* เพิ่ม onClick และ style ตรงนี้ */}
                      <div
                        className="friends-left"
                        onClick={() => goToProfile(r.requester_user_id)}
                        style={{ cursor: "pointer" }}
                        title="ไปที่โปรไฟล์"
                      >
                        <img
                          className="friends-avatar"
                          src={toAbsUrl(r.avatar) || picdefault}
                          alt={`req-${r.request_id}`}
                          onError={(e) => (e.currentTarget.src = picdefault)}
                        />
                        <div className="friends-name">
                          <span className="friends-name-main">
                            {r.username || `user#${r.requester_user_id}`}
                          </span>
                        </div>
                      </div>

                      <div className="friends-actions-right">
                        <button
                          className="friends-pill friends-pill--green"
                          onClick={() => acceptRequest(r.request_id)}
                        >
                          ยอมรับ
                        </button>
                        <button
                          className="friends-pill friends-pill--danger"
                          onClick={() => declineRequest(r.request_id)}
                        >
                          ปฏิเสธ
                        </button>
                      </div>
                    </li>
                  ))}

                  {incoming.length === 0 && (
                    <div className="friends-placeholder">
                      ยังไม่มีคำขอเข้ามา
                    </div>
                  )}
                </ul>
              )}
            </>
          )}
        </main>
      </div>

      <Footer />
    </div>
  );
};

export default Friends;
