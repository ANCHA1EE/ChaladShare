import React, { useEffect, useMemo, useState } from "react";
import { useParams, Link, useNavigate } from "react-router-dom";
import { useNotification } from "../component/Notification";
import { VscEye, VscEyeClosed } from "react-icons/vsc";
import { FiEdit2, FiTrash2 } from "react-icons/fi";
import axios from "axios";
import Sidebar from "./Sidebar";
import Footer from "../component/Footer";
import PostCard from "../component/Postcard";
import Avatar from "../assets/default.png";
import "../component/Profile.css";

const API_HOST = "http://localhost:8080";
const toAbsUrl = (p) => {
  if (!p) return "";
  if (p.startsWith("http")) return p;
  const clean = p.replace(/^\./, "");
  return `${API_HOST}${clean.startsWith("/") ? clean : `/${clean}`}`;
};

const Profile = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const { success: notifySuccess, error: notifyError } = useNotification();

  const [myId, setMyId] = useState(null);
  const isOwn = useMemo(
    () => (myId == null ? null : !id || String(id) === String(myId)),
    [id, myId],
  );
  const targetId = id || null;
  const ownerId = useMemo(() => {
    if (isOwn == null) return null;
    return isOwn ? myId : targetId;
  }, [isOwn, myId, targetId]);

  const [activeTab, setActiveTab] = useState("posts");
  const [profile, setProfile] = useState({
    name: "",
    bio: "",
    email: "",
    avatar: Avatar,
    posts: 0,
    followers: 0,
    following: 0,
  });

  // โหมดแก้ไขโปรไฟล์
  const [isEditing, setIsEditing] = useState(false);
  const [avatarPreview, setAvatarPreview] = useState(null);
  const [avatarFile, setAvatarFile] = useState(null);
  const [editForm, setEditForm] = useState({
    name: "",
    email: "",
    bio: "",
  });
  const [pwdForm, setPwdForm] = useState({
    current: "",
    newPwd: "",
    confirm: "",
  });
  const [saving, setSaving] = useState(false);
  const [saveErr, setSaveErr] = useState("");

  // state เปิด/ปิดการมองรหัสผ่าน
  const [showPwd, setShowPwd] = useState({
    current: false,
    newPwd: false,
    confirm: false,
  });
  const toggleShow = (key) => setShowPwd((s) => ({ ...s, [key]: !s[key] }));

  const openEdit = () => {
    setEditForm({
      name: profile.name || "",
      email: profile.email || "",
      bio: profile.bio || "",
    });
    setPwdForm({ current: "", newPwd: "", confirm: "" });
    setAvatarPreview(null);
    setAvatarFile(null);
    setSaveErr("");
    setIsEditing(true);
  };

  const onPickAvatar = (e) => {
    const f = e.target.files?.[0];
    if (!f) return;
    setAvatarFile(f);
    const url = URL.createObjectURL(f);
    setAvatarPreview(url);
  };

  const submitEdit = async () => {
    try {
      setSaving(true);
      setSaveErr("");

      // validate เปลี่ยนรหัสผ่าน (แก้เฉพาะที่จำเป็น)
      const hasAnyPwd =
        !!pwdForm.current.trim() ||
        !!pwdForm.newPwd.trim() ||
        !!pwdForm.confirm.trim();

      if (hasAnyPwd) {
        if (
          !pwdForm.current.trim() ||
          !pwdForm.newPwd.trim() ||
          !pwdForm.confirm.trim()
        ) {
          setSaveErr(
            "หากต้องการเปลี่ยนรหัสผ่าน กรุณากรอก รหัสผ่านปัจจุบัน / รหัสผ่านใหม่ / ยืนยันรหัสผ่านใหม่ ให้ครบ",
          );
          setSaving(false);
          return;
        }
        if (pwdForm.newPwd.length < 8) {
          setSaveErr("รหัสผ่านใหม่ต้องมีอย่างน้อย 8 ตัวอักษร");
          setSaving(false);
          return;
        }
        if (pwdForm.newPwd !== pwdForm.confirm) {
          setSaveErr("ยืนยันรหัสผ่านใหม่ไม่ตรงกัน");
          setSaving(false);
          return;
        }
      }

      // อัปโหลดรูป
      let avatarUrl = null;
      let avatarStorage = null;

      if (avatarFile) {
        const fd = new FormData();
        fd.append("file", avatarFile);
        const res = await axios.post("/files/avatar", fd, {
          headers: { "Content-Type": "multipart/form-data" },
        });

        avatarUrl = res?.data?.avatar_url || null;
        avatarStorage = res?.data?.avatar_storage || "local";
      }

      // edit profile
      await axios.put("/profile", {
        username: editForm.name,
        bio: editForm.bio,
        ...(avatarUrl && {
          avatar_url: avatarUrl,
          avatar_storage: avatarStorage,
        }),
      });

      if (pwdForm.current && pwdForm.newPwd && pwdForm.confirm) {
        await axios.post("/profile/change-password", {
          current_password: pwdForm.current,
          new_password: pwdForm.newPwd,
          confirm_password: pwdForm.confirm,
        });
      }

      // อัปเดตสถานะบนหน้า
      const fullAvatarUrl = avatarPreview
        ? avatarPreview
        : avatarUrl
          ? toAbsUrl(avatarUrl)
          : profile.avatar;

      setProfile((p) => ({
        ...p,
        name: editForm.name,
        email: editForm.email,
        bio: editForm.bio,
        avatar: avatarPreview || fullAvatarUrl || p.avatar,
      }));
      notifySuccess("แก้ไขโปรไฟล์สำเร็จ", 2500);

      setIsEditing(false);
    } catch (e) {
      const raw =
        e?.response?.data?.error ||
        e?.response?.data?.message ||
        e?.message ||
        "บันทึกล้มเหลว";

      const msg =
        String(raw).includes("users_username_key") ||
        String(raw).toLowerCase().includes("duplicate key") ||
        String(raw).includes("23505")
          ? "ชื่อผู้ใช้นี้ถูกใช้งานแล้ว"
          : raw;

      setSaveErr(msg);
    } finally {
      setSaving(false);
    }
  };

  const [posts, setPosts] = useState([]);
  const [savedPosts, setSavedPosts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState("");
  const [followStatus, setFollowStatus] = useState("idle");
  const [friendStatus, setFriendStatus] = useState("idle");

  // dropdown ของปุ่มเพื่อน
  const [friendMenuOpen, setFriendMenuOpen] = useState(false);

  // ปิด dropdown เมื่อคลิกข้างนอก
  useEffect(() => {
    const onDocClick = (e) => {
      if (!e.target.closest(".friend-dropdown-wrap")) {
        setFriendMenuOpen(false);
      }
    };
    document.addEventListener("click", onDocClick);
    return () => document.removeEventListener("click", onDocClick);
  }, []);

  // ปิดเมนูเมื่อเปลี่ยนโปรไฟล์/สถานะ
  useEffect(() => {
    setFriendMenuOpen(false);
  }, [targetId, friendStatus]);

  const goToPostDetail = (post) => {
    if (post?.id) navigate(`/posts/${post.id}`);
  };

  // แก้ไข ลบโพสต์
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState(null); // เก็บ post ที่จะลบ
  const [deleting, setDeleting] = useState(false);

  const handleEditPost = (event, post) => {
    event.stopPropagation();
    if (!post?.id) return;
    navigate(`/posts/${post.id}/edit`);
  };

  // เปลี่ยนจากลบทันที -> เปิด popup ยืนยัน
  const handleDeletePost = (event, post) => {
    event.stopPropagation();
    if (!post?.id) return;
    setDeleteTarget(post);
    setDeleteOpen(true);
  };

  // กดลบโพสต์ใน popup ถึงจะลบจริง
  const confirmDelete = async () => {
    if (!deleteTarget?.id) return;

    try {
      setDeleting(true);

      await axios.delete(`/posts/${deleteTarget.id}`);

      setPosts((list) => list.filter((p) => p.id !== deleteTarget.id));
      setProfile((p) => ({
        ...p,
        posts: Math.max(0, (p.posts ?? 0) - 1),
      }));

      setDeleteOpen(false);
      setDeleteTarget(null);

      notifySuccess("ลบโพสต์สำเร็จ", 2500);
    } catch (e) {
      const msg = e?.response?.data?.error || "ลบโพสต์ไม่สำเร็จ";
      notifyError(msg, 2500);
    } finally {
      setDeleting(false);
    }
  };

  // โหลด myId
  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const me = await axios.get("/profile");
        if (!cancelled) setMyId(me?.data?.user_id ?? me?.data?.id ?? null);
      } catch (e) {
        if (e?.response?.status === 401) navigate("/", { replace: true });
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [navigate]);

  useEffect(() => {
    if (isOwn == null || !ownerId) return;
    setLoading(true);
    setErr("");

    const fetchData = async () => {
      try {
        const prof = isOwn
          ? await axios.get("/profile", {
              params: { with: "stats,followers,following,rel" },
            })
          : await axios.get(`/profile/${targetId}`, {
              params: { with: "stats,followers,following,rel" },
            });

        const statsUserId = isOwn
          ? (prof?.data?.user_id ?? prof?.data?.id ?? myId)
          : ownerId;
        const statsRes = await axios.get(`/social/stats/${statsUserId}`);
        const stats = statsRes?.data ?? {};

        const postsRes = isOwn
          ? await axios.get("/posts", { params: { mine: 1 } })
          : await axios.get("/posts", { params: { user_id: ownerId } });

        let savedRes = { data: [] };
        if (isOwn) {
          try {
            savedRes = await axios.get("/posts/save");
          } catch {}
        }

        const rawAvatar = prof?.data?.avatar_url || "";

        const format = (list) =>
          Array.isArray(list)
            ? list.map((p) => {
                const fileRaw = p.file_url || "";
                const coverRaw = p.cover_url || "";
                const isPdf = /\.pdf$/i.test(fileRaw);

                const imgSrc = coverRaw
                  ? toAbsUrl(coverRaw)
                  : !fileRaw || isPdf
                    ? "/img/pdf-placeholder.jpg"
                    : toAbsUrl(fileRaw);

                const rawAuthorAvatar = p.avatar_url || "";
                const authorImg = rawAuthorAvatar
                  ? toAbsUrl(rawAuthorAvatar)
                  : Avatar;

                const authorName =
                  p.author_name ||
                  p.username ||
                  (isOwn && profile.name) ||
                  "ผู้ใช้";

                return {
                  id: p.post_id,
                  post: p.post_id,
                  img: imgSrc,
                  isPdf,
                  likes: p.like_count ?? 0,
                  like_count: p.like_count ?? 0,
                  is_liked: !!p.is_liked,
                  is_saved: !!p.is_saved,
                  title: p.post_title,
                  tags: Array.isArray(p.tags)
                    ? p.tags
                        .map((t) => (t.startsWith("#") ? t : `#${t}`))
                        .join(" ")
                    : "",
                  authorId: p.author_id ?? p.post_user_id ?? p.user_id,
                  authorName,
                  authorImg,
                };
              })
            : [];

        setProfile((prev) => {
          const avatarFull = rawAvatar
            ? toAbsUrl(rawAvatar)
            : prev.avatar || Avatar;
          return {
            ...prev,
            name: prof?.data?.username ?? prev.name,
            email: prof?.data?.email ?? prev.email,
            bio: prof?.data?.bio ?? prev.bio,
            avatar: avatarFull,
            posts: prof?.data?.posts_count ?? prev.posts ?? 0,
            followers: stats.followers ?? prev.followers ?? 0,
            following: stats.following ?? prev.following ?? 0,
          };
        });

        const postRows = Array.isArray(postsRes?.data?.data)
          ? postsRes.data.data
          : Array.isArray(postsRes?.data)
            ? postsRes.data
            : [];
        const savedRows = Array.isArray(savedRes?.data?.data)
          ? savedRes.data.data
          : Array.isArray(savedRes?.data)
            ? savedRes.data
            : [];

        const ownerRows = postRows.filter(
          (p) =>
            String(p.author_id ?? p.post_user_id ?? p.user_id) ===
            String(ownerId),
        );

        setPosts(format(ownerRows));
        setSavedPosts(isOwn ? format(savedRows) : []);

        if (!isOwn) {
          const rel = prof?.data ?? {};

          // follow
          if (typeof rel.is_following === "boolean") {
            setFollowStatus(rel.is_following ? "following" : "idle");
          } else if (rel.is_following) {
            setFollowStatus("following");
          } else {
            setFollowStatus("idle");
          }

          // friend
          if (
            rel.is_friend === true ||
            rel.is_friend === 1 ||
            rel.is_friend === "1"
          ) {
            setFriendStatus("friends");
          } else if (rel.friend_request_outgoing) {
            setFriendStatus("requested");
          } else if (rel.friend_request_incoming) {
            setFriendStatus("incoming");
          } else {
            setFriendStatus("idle");
          }
        } else {
          setFollowStatus("idle");
          setFriendStatus("idle");
        }
      } catch (e) {
        setErr(e?.response?.data?.error || e.message);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [isOwn, ownerId, targetId, myId, profile.name]);

  useEffect(() => {
    if (isOwn || !myId || !targetId) return;
    if (friendStatus === "friends") return;

    const checkFriendRelation = async () => {
      try {
        const friendsRes = await axios.get(`/social/friends/${myId}`, {
          params: { page: 1, size: 500, search: "" },
        });
        const friendItems = friendsRes.data.items || [];
        const isFriend = friendItems.some(
          (f) => String(f.user_id) === String(targetId),
        );

        if (isFriend) {
          setFriendStatus("friends");
          return;
        }

        const outgoingRes = await axios.get("/social/requests/outgoing", {
          params: { page: 1, size: 500 },
        });
        const outgoingItems = outgoingRes.data.items || [];

        const hasOutgoing = outgoingItems.some((r) => {
          const toId =
            r.addressee_user_id ??
            r.addressee_id ??
            r.to_user_id ??
            r.receiver_id ??
            r.requested_user_id ??
            r.target_user_id;

          return String(toId) === String(targetId);
        });

        const incomingRes = await axios.get("/social/requests/incoming", {
          params: { page: 1, size: 500 },
        });
        const incomingItems = incomingRes.data.items || [];

        const hasIncoming = incomingItems.some((r) => {
          const fromId =
            r.requester_user_id ??
            r.requester_id ??
            r.from_user_id ??
            r.sender_id ??
            r.user_id;

          return String(fromId) === String(targetId);
        });

        if (hasOutgoing) setFriendStatus("requested");
        else if (hasIncoming) setFriendStatus("incoming");
        else setFriendStatus("idle");
      } catch (e) {
        console.error(
          "checkFriendRelation failed:",
          e?.response?.status,
          e?.response?.data || e,
        );
      }
    };

    checkFriendRelation();
  }, [isOwn, myId, targetId, friendStatus]);

  useEffect(() => {
    if (isOwn === false && activeTab !== "posts") setActiveTab("posts");
  }, [isOwn, activeTab]);

  const onToggleFollow = async () => {
    if (isOwn || !targetId) return;
    try {
      if (followStatus === "following") {
        await axios.delete(`/social/follow/${targetId}`);
        setFollowStatus("idle");
        setProfile((p) => ({
          ...p,
          followers: Math.max(0, (p.followers ?? 0) - 1),
        }));
      } else {
        await axios.post(`/social/follow`, {
          followed_user_id: Number(targetId),
        });
        setFollowStatus("following");
        setProfile((p) => ({ ...p, followers: (p.followers ?? 0) + 1 }));
      }
    } catch (e) {
      console.error(e);
    }
  };

  const onAddFriend = async () => {
    if (isOwn || !targetId || friendStatus !== "idle") return;
    try {
      await axios.post("/social/requests", { to_user_id: Number(targetId) });
      setFriendStatus("requested");
    } catch (e) {
      console.error(e);
    }
  };

  const onAcceptFriendRequest = async () => {
    if (isOwn || !targetId) return;

    try {
      const incomingRes = await axios.get("/social/requests/incoming", {
        params: { page: 1, size: 500 },
      });
      const items = incomingRes?.data?.items || [];

      const req = items.find((r) => {
        const fromId =
          r.requester_user_id ??
          r.requester_id ??
          r.from_user_id ??
          r.sender_id ??
          r.user_id;

        return String(fromId) === String(targetId);
      });

      if (!req) {
        setFriendStatus("idle");
        setFriendMenuOpen(false);
        return;
      }

      const requestId = req.request_id ?? req.friend_request_id ?? req.id;
      if (!requestId) return;

      await axios.post(`/social/requests/${requestId}/accept`);
      setFriendStatus("friends");
      setFriendMenuOpen(false);
    } catch (e) {
      console.error(e);
    }
  };

  const onRejectFriendRequest = async () => {
    if (isOwn || !targetId) return;

    try {
      const incomingRes = await axios.get("/social/requests/incoming", {
        params: { page: 1, size: 500 },
      });
      const items = incomingRes?.data?.items || [];

      const req = items.find((r) => {
        const fromId =
          r.requester_user_id ??
          r.requester_id ??
          r.from_user_id ??
          r.sender_id ??
          r.user_id;

        return String(fromId) === String(targetId);
      });

      if (!req) {
        setFriendStatus("idle");
        setFriendMenuOpen(false);
        return;
      }

      const requestId = req.request_id ?? req.friend_request_id ?? req.id;
      if (!requestId) return;

      await axios.post(`/social/requests/${requestId}/reject`);
      setFriendStatus("idle");
      setFriendMenuOpen(false);
    } catch (e) {
      console.error(e);
    }
  };

  const onCancelFriendRequest = async () => {
    if (isOwn || !targetId) return;

    try {
      const outgoingRes = await axios.get("/social/requests/outgoing", {
        params: { page: 1, size: 500 },
      });
      const items = outgoingRes?.data?.items || [];

      const req = items.find((r) => {
        const toId =
          r.addressee_user_id ??
          r.addressee_id ??
          r.to_user_id ??
          r.receiver_id ??
          r.requested_user_id ??
          r.target_user_id;

        return String(toId) === String(targetId);
      });

      if (!req) {
        setFriendStatus("idle");
        setFriendMenuOpen(false);
        return;
      }

      const requestId = req.request_id ?? req.friend_request_id ?? req.id;
      if (!requestId) {
        console.warn(
          "No request id field found in outgoing request item:",
          req,
        );
        return;
      }

      await axios.delete(`/social/requests/${requestId}`);
      setFriendStatus("idle");
      setFriendMenuOpen(false);
    } catch (e) {
      console.error(e);
    }
  };

  const onUnfriend = async () => {
    if (isOwn || !targetId) return;

    try {
      await axios.delete(`/social/friends/${targetId}`);
      setFriendStatus("idle");
      setFriendMenuOpen(false);
    } catch (e) {
      console.error(e);
    }
  };

  const showing = useMemo(
    () => (activeTab === "posts" ? posts : savedPosts),
    [activeTab, posts, savedPosts],
  );

  if (isOwn == null) {
    return (
      <div className="profile-page">
        <div className="profile-container">
          <Sidebar />
          <main className="profile-content">
            <div className="profile-shell">
              <p className="profile-msg">กำลังโหลด...</p>
            </div>
          </main>
        </div>
      </div>
    );
  }

  return (
    <div className="profile-page">
      <div className="profile-container">
        <Sidebar />

        <main className="profile-content">
          <div className="profile-shell">
            {/* Header */}
            <section className="profile-header">
              <img
                src={profile.avatar}
                alt={profile.name}
                className="profile-avatar"
              />
              <div className="profile-info">
                <div className="profile-toprow">
                  {isOwn ? (
                    <h2 className="profile-name">{profile.name || "—"}</h2>
                  ) : (
                    <h2 className="profile-name">
                      <Link to={`/profile/${targetId}`}>
                        {profile.name || "—"}
                      </Link>
                    </h2>
                  )}

                  {isOwn ? (
                    <button className="profile-btn-edit" onClick={openEdit}>
                      แก้ไขโปรไฟล์
                    </button>
                  ) : (
                    <div className="profile-actions">
                      <button
                        className={`btn-follow ${followStatus === "following" ? "is-following" : ""}`}
                        onClick={onToggleFollow}
                      >
                        {followStatus === "following"
                          ? "กำลังติดตาม"
                          : "ติดตาม"}
                      </button>

                      <div className="friend-dropdown-wrap">
                        <button
                          className={`btn-friend ${
                            friendStatus === "friends"
                              ? "is-friends"
                              : friendStatus === "requested"
                                ? "is-requested"
                                : friendStatus === "incoming"
                                  ? "is-incoming"
                                  : ""
                          }`}
                          onClick={() => {
                            if (friendStatus === "idle") onAddFriend();
                            else setFriendMenuOpen((v) => !v);
                          }}
                        >
                          {friendStatus === "friends"
                            ? "เพื่อน"
                            : friendStatus === "requested"
                              ? "กำลังรอการตอบรับ"
                              : friendStatus === "incoming"
                                ? "มีคำขอเป็นเพื่อน"
                                : "+ เพิ่มเพื่อน"}
                        </button>

                        {(friendStatus === "requested" ||
                          friendStatus === "friends" ||
                          friendStatus === "incoming") &&
                          friendMenuOpen && (
                            <div className="friend-dropdown">
                              {friendStatus === "requested" && (
                                <button
                                  type="button"
                                  className="friend-dropdown-item danger"
                                  onClick={onCancelFriendRequest}
                                >
                                  ยกเลิกคำขอ
                                </button>
                              )}

                              {friendStatus === "incoming" && (
                                <>
                                  <button
                                    type="button"
                                    className="friend-dropdown-item"
                                    onClick={onAcceptFriendRequest}
                                  >
                                    ยอมรับคำขอ
                                  </button>
                                  <button
                                    type="button"
                                    className="friend-dropdown-item danger"
                                    onClick={onRejectFriendRequest}
                                  >
                                    ปฏิเสธคำขอ
                                  </button>
                                </>
                              )}

                              {friendStatus === "friends" && (
                                <button
                                  type="button"
                                  className="friend-dropdown-item danger"
                                  onClick={onUnfriend}
                                >
                                  เลิกเป็นเพื่อน
                                </button>
                              )}
                            </div>
                          )}
                      </div>
                    </div>
                  )}
                </div>

                <p className="profile-bio">{profile.bio || ""}</p>
                <div className="profile-stats">
                  <span>{profile.posts} โพสต์</span>
                  <span>{profile.followers} ผู้ติดตาม</span>
                  <span>{profile.following} กำลังติดตาม</span>
                </div>
              </div>
            </section>

            {/* Edit Card */}
            {isOwn && isEditing && (
              <section className="edit-card">
                <div className="edit-grid">
                  <div className="edit-avatar-col">
                    <img
                      src={avatarPreview || profile.avatar}
                      alt="avatar-preview"
                      className="edit-avatar"
                    />
                    <label className="edit-upload-btn">
                      เปลี่ยนรูปโปรไฟล์
                      <input
                        type="file"
                        accept="image/*"
                        onChange={onPickAvatar}
                        hidden
                      />
                    </label>
                  </div>

                  <div className="edit-form-col">
                    <div className="edit-field">
                      <label>ชื่อผู้ใช้</label>
                      <input
                        type="text"
                        value={editForm.name}
                        onChange={(e) =>
                          setEditForm((f) => ({ ...f, name: e.target.value }))
                        }
                      />
                    </div>

                    <div className="edit-field">
                      <label>อีเมล</label>
                      <input type="email" value={editForm.email} readOnly />
                    </div>

                    <div className="edit-field">
                      <label>คำอธิบาย</label>
                      <textarea
                        rows={3}
                        value={editForm.bio}
                        onChange={(e) =>
                          setEditForm((f) => ({ ...f, bio: e.target.value }))
                        }
                      />
                    </div>

                    {/* แถวเดียว: รหัสผ่านปัจจุบัน + ลิงก์ลืมรหัสผ่าน */}
                    <div className="edit-row">
                      <div className="edit-field flex-1">
                        <label>รหัสผ่านปัจจุบัน</label>
                        <div className="pw-field">
                          <input
                            className="pw-input"
                            type={showPwd.current ? "text" : "password"}
                            value={pwdForm.current}
                            onChange={(e) =>
                              setPwdForm((p) => ({
                                ...p,
                                current: e.target.value,
                              }))
                            }
                            placeholder="กรอกรหัสผ่านปัจจุบัน"
                          />
                          <button
                            type="button"
                            className="pw-toggle"
                            onClick={() => toggleShow("current")}
                            aria-label={
                              showPwd.current ? "ซ่อนรหัสผ่าน" : "แสดงรหัสผ่าน"
                            }
                            title={
                              showPwd.current ? "ซ่อนรหัสผ่าน" : "แสดงรหัสผ่าน"
                            }
                          >
                            {showPwd.current ? <VscEyeClosed /> : <VscEye />}
                          </button>
                        </div>
                      </div>

                      <button
                        type="button"
                        className="forgot-link"
                        onClick={() => navigate("/forgot_password")}
                        aria-label="ไปหน้าลืมรหัสผ่าน"
                        title="ลืมรหัสผ่าน?"
                      >
                        ลืมรหัสผ่าน?
                      </button>
                    </div>

                    {/* แถว: รหัสผ่านใหม่ / ยืนยันรหัสผ่านใหม่ */}
                    <div className="edit-2col">
                      <div className="edit-field">
                        <label>รหัสผ่านใหม่</label>
                        <div className="pw-field">
                          <input
                            className="pw-input"
                            type={showPwd.newPwd ? "text" : "password"}
                            value={pwdForm.newPwd}
                            onChange={(e) =>
                              setPwdForm((p) => ({
                                ...p,
                                newPwd: e.target.value,
                              }))
                            }
                            placeholder="อย่างน้อย 8 ตัวอักษร"
                          />
                          <button
                            type="button"
                            className="pw-toggle"
                            onClick={() => toggleShow("newPwd")}
                            aria-label={
                              showPwd.newPwd ? "ซ่อนรหัสผ่าน" : "แสดงรหัสผ่าน"
                            }
                          >
                            {showPwd.newPwd ? <VscEyeClosed /> : <VscEye />}
                          </button>
                        </div>
                      </div>

                      <div className="edit-field">
                        <label>ยืนยันรหัสผ่านใหม่</label>
                        <div className="pw-field">
                          <input
                            className="pw-input"
                            type={showPwd.confirm ? "text" : "password"}
                            value={pwdForm.confirm}
                            onChange={(e) =>
                              setPwdForm((p) => ({
                                ...p,
                                confirm: e.target.value,
                              }))
                            }
                            placeholder="พิมพ์ให้ตรงกับรหัสใหม่"
                          />
                          <button
                            type="button"
                            className="pw-toggle"
                            onClick={() => toggleShow("confirm")}
                            aria-label={
                              showPwd.confirm ? "ซ่อนรหัสผ่าน" : "แสดงรหัสผ่าน"
                            }
                          >
                            {showPwd.confirm ? <VscEyeClosed /> : <VscEye />}
                          </button>
                        </div>
                      </div>
                    </div>

                    {saveErr && <p className="edit-error">{saveErr}</p>}

                    <div className="edit-actions">
                      <button
                        className="btn-cancel"
                        onClick={() => setIsEditing(false)}
                        disabled={saving}
                      >
                        ยกเลิก
                      </button>
                      <button
                        className="btn-save"
                        onClick={submitEdit}
                        disabled={saving}
                      >
                        {saving ? "กำลังบันทึก…" : "บันทึกการแก้ไข"}
                      </button>
                    </div>
                  </div>
                </div>
              </section>
            )}
            {/* Edit*/}

            {/* Tabs */}
            <div className="profile-tabs">
              <div className="profile-tabs-track">
                <button
                  className={`profile-tab ${activeTab === "posts" ? "active" : ""}`}
                  onClick={() => setActiveTab("posts")}
                >
                  โพสต์
                </button>
                {isOwn && (
                  <button
                    className={`profile-tab ${activeTab === "saved" ? "active" : ""}`}
                    onClick={() => setActiveTab("saved")}
                  >
                    รายการที่บันทึก
                  </button>
                )}
                <div
                  className={`profile-tab-underline ${activeTab === "saved" ? "saved" : ""}`}
                />
              </div>
            </div>

            {/* Posts */}
            <section className="profile-body">
              {loading ? (
                <p className="profile-msg">กำลังโหลด...</p>
              ) : err ? (
                <p className="profile-msg error">เกิดข้อผิดพลาด: {err}</p>
              ) : showing.length === 0 ? (
                <p className="profile-msg muted">ไม่มีโพสต์ที่จะแสดง</p>
              ) : (
                <div className="card-list">
                  {showing.map((post) => (
                    <div
                      key={post.id}
                      className="profile-card-wrapper"
                      onClick={() => goToPostDetail(post)}
                      style={{ cursor: "pointer" }}
                    >
                      <PostCard post={post} />

                      {isOwn && activeTab === "posts" && (
                        <div className="profile-manage-row">
                          <button
                            type="button"
                            className="profile-manage-btn profile-manage-btn-edit"
                            onClick={(e) => handleEditPost(e, post)}
                          >
                            <FiEdit2 />
                            แก้ไข
                          </button>
                          <button
                            type="button"
                            className="profile-manage-btn profile-manage-btn-delete"
                            onClick={(e) => handleDeletePost(e, post)}
                          >
                            <FiTrash2 />
                            ลบ
                          </button>
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </section>

            {deleteOpen && (
              <div
                className="modal-overlay"
                onClick={() => {
                  if (deleting) return;
                  setDeleteOpen(false);
                  setDeleteTarget(null);
                }}
              >
                <div
                  className="modal-card"
                  role="dialog"
                  aria-modal="true"
                  onClick={(e) => e.stopPropagation()}
                >
                  <h3 className="modal-title">ยืนยันการลบโพสต์</h3>
                  <p className="modal-desc">
                    คุณต้องการลบโพสต์ <b>{deleteTarget?.title || ""}</b> ใช่ไหม?
                  </p>

                  <div className="modal-actions">
                    <button
                      className="modal-btn"
                      onClick={() => {
                        setDeleteOpen(false);
                        setDeleteTarget(null);
                      }}
                      disabled={deleting}
                    >
                      ยกเลิก
                    </button>

                    <button
                      className="modal-btn danger"
                      onClick={confirmDelete}
                      disabled={deleting}
                    >
                      {deleting ? "กำลังลบ..." : "ลบโพสต์"}
                    </button>
                  </div>
                </div>
              </div>
            )}
          </div>
        </main>
      </div>
      <Footer />
    </div>
  );
};

export default Profile;
