import React, { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { FaArrowLeft } from "react-icons/fa";
import { AiFillHeart, AiOutlineHeart } from "react-icons/ai";
import { FiShare2 } from "react-icons/fi";
import { BsBookmark, BsBookmarkFill } from "react-icons/bs";
import axios from "axios";
import Sidebar from "./Sidebar";
import Avatar from "../assets/default.png";
import "../component/PostDetail.css";

const API_HOST = "http://localhost:8080";
// const API_ORIGIN =
//   process.env.REACT_APP_API_ORIGIN || window.location.origin;

const toAbsUrl = (p) => {
  if (!p) return "";
  if (p.startsWith("http")) return p;
  const clean = p.replace(/^\./, "");
  return `${API_HOST}${clean.startsWith("/") ? clean : `/${clean}`}`;
  // return `${API_ORIGIN}${path}`;
};

const PostDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();

  const [post, setPost] = useState(null);
  const [liked, setLiked] = useState(false);
  const [likes, setLikes] = useState(0);
  const [saved, setSaved] = useState(false);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState("");

  useEffect(() => {
  (async () => {
    try {
      setLoading(true);
      setErr("");

      const res = await axios.get(`/posts/${id}`, { withCredentials: true });
      const payload = res?.data?.data ?? res?.data ?? {};
      const data = payload?.post ?? payload ?? {};

      if (!data || (!data.post_id && !data.id)) {
        setErr("ไม่พบโพสต์");
        return;
      }

      const avatarRaw = data.avatar_url || data.AvatarURL || "";
      const authorImg = avatarRaw ? toAbsUrl(avatarRaw) : Avatar;

      const rawFile =
        data.file_url || data.document_url || data.document_file_url || "";

      const mapped = {
        id: data.post_id ?? data.id,
        title: data.post_title ?? data.title ?? "",
        description: data.post_description ?? data.description ?? "",
        visibility: data.post_visibility ?? data.visibility ?? "public",
        file_url: rawFile ? toAbsUrl(rawFile) : null,

        author_name: data.author_name ?? data.authorName ?? "",
        author_id: data.author_id ?? data.author_user_id ?? data.post_author_user_id,

        authorImg,

        like_count: data.like_count ?? data.post_like_count ?? 0,
        is_liked: !!(data.is_liked ?? data.isLiked),
        is_saved: !!(data.is_saved ?? data.isSaved),

        tags: data.tags || [],
        post_document_id: data.post_document_id ?? data.document_id,
      };

      setPost(mapped);
      setLikes(mapped.like_count || 0);
      setLiked(!!mapped.is_liked);
      setSaved(!!mapped.is_saved);
    } catch (e) {
      const st = e?.response?.status;
      if (st === 401) {
        navigate("/", { replace: true });
        return;
      }
      if (st === 403) setErr("คุณไม่มีสิทธิ์ดูโพสต์นี้");
      else if (st === 404) setErr("ไม่พบโพสต์");
      else setErr(e?.response?.data?.error || e.message || "โหลดโพสต์ล้มเหลว");
    } finally {
      setLoading(false);
    }
  })();
}, [id, navigate]);


  // Like แบบเดียวกับการ์ด Home (optimistic + ไม่ revert UI)
  const toggleLike = async (e) => {
    e?.preventDefault?.();
    e?.stopPropagation?.();

    if (!post) return;

  try {
    const res = await axios.post(
      `/posts/${id}/like`,
      {},
      { withCredentials: true }
    );
    const { is_liked, like_count } = res.data.data;

    setLiked(is_liked);
    setLikes(like_count ?? 0);
    setPost((prev) =>
      prev ? { ...prev, is_liked, like_count: like_count ?? 0 } : prev
    );
  } catch (error) {
    console.error("Like toggle failed:", error);
  }
};

  // Save แบบเดียวกับ Home (ไอคอนเหลืองเมื่อ active)
  const toggleSave = async (e) => {
    e?.preventDefault?.();
    e?.stopPropagation?.();

    if (!post) return;

  try {
    const res = await axios.post(
      `/posts/${id}/save`,
      {},
      { withCredentials: true }
    );
    const { is_saved, save_count } = res.data.data;

    setSaved(is_saved);
    setPost((prev) =>
      prev ? { ...prev, is_saved, save_count: save_count ?? prev.save_count } : prev
    );
  } catch (error) {
    console.error("Save toggle failed:", error);
  }
};

  const sharePost = async (e) => {
    e?.preventDefault?.();
    e?.stopPropagation?.();
    const url = window.location.href;
    try {
      if (navigator.share) {
        await navigator.share({
          title: post?.title || "ChaladShare",
          text: post?.description || "",
          url,
        });
      } else {
        await navigator.clipboard.writeText(url);
        alert("คัดลอกลิงก์แล้ว");
      }
    } catch {
      /* ผู้ใช้ยกเลิกการแชร์ */
    }
  };

  if (loading)
    return (
      <div className="post-detail-page">
        <Sidebar />
        <main className="post-detail">
          <div style={{ padding: 24 }}>กำลังโหลด…</div>
        </main>
      </div>
    );

  if (err)
    return (
      <div className="post-detail-page">
        <Sidebar />
        <main className="post-detail">
          <div style={{ padding: 24, color: "#b00020" }}>{err}</div>
        </main>
      </div>
    );

  if (!post)
    return (
      <div className="post-detail-page">
        <Sidebar />
        <main className="post-detail">
          <div style={{ padding: 24 }}>ไม่พบโพสต์</div>
        </main>
      </div>
    );

  const isPdf =
    Boolean(post.post_document_id) || /\.pdf$/i.test(post.file_url || "");
  const visibilityText =
    post.visibility === "friends" ? "เฉพาะเพื่อน" : "สาธารณะ";

  return (
    <div className="post-detail-page">
      <Sidebar />

      <main className="post-detail">
        {/* Header */}
        <header className="post-header">
          <button
            type="button"
            className="back-btn"
            onClick={() => navigate(-1)}
            aria-label="ย้อนกลับ"
          >
            <FaArrowLeft />
          </button>

          <div
            className="user-info"
            style={{ cursor: post.author_id ? "pointer" : "default" }}
            onClick={() =>
              post.author_id && navigate(`/profile/${post.author_id}`)
            }
            title={post.author_id ? "ไปที่โปรไฟล์ผู้เขียน" : undefined}
          >
            <img src={post.authorImg} alt="profile" className="profile-img" />
            <div className="user-details">
              <h4>{post.author_name || "ไม่ระบุ"}</h4>
              <p className="status">{visibilityText}</p>
            </div>
          </div>
        </header>

        {/* Viewer */}
        <section className="post-image">
          <div className="pdf-slide-wrapper">
            {post.file_url ? (
              isPdf ? (
                <iframe
                  className="pdf-frame"
                  src={`${post.file_url}#zoom=page-width`}
                  title="pdf"
                />
              ) : (
                <img
                  className="pdf-page-img active"
                  src={post.file_url}
                  alt={post.title}
                />
              )
            ) : (
              <img
                className="pdf-page-img active"
                src="/img/no-image.png"
                alt={post.title}
              />
            )}
          </div>
        </section>

        {/* Actions + Meta */}
        <section className="post-footer">
          <div
            className="actions-row"
            onClick={(e) => e.stopPropagation()} // กัน bubble ออกไปนอกแถว
          >
            {/* ไลก์สไตล์เดียวกับการ์ด Home */}
            <span
              className="likes"
              onClick={toggleLike}
              style={{ cursor: "pointer" }}
            >
              {liked ? (
                <AiFillHeart style={{ color: "red", fontSize: "20px" }} />
              ) : (
                <AiOutlineHeart style={{ color: "black", fontSize: "20px" }} />
              )}
              <span>{likes}</span>
            </span>

            {/* ปุ่มขวา: บันทึก + แชร์ */}
            <div className="action-right">
              <button
                type="button"
                className={`icon-btn ${saved ? "active" : ""}`}
                onClick={toggleSave}
                title={saved ? "ยกเลิกบันทึก" : "บันทึก"}
                aria-label="บันทึก"
              >
                {saved ? <BsBookmarkFill /> : <BsBookmark />}
              </button>

              <button
                type="button"
                className="icon-btn"
                onClick={sharePost}
                title="แชร์"
                aria-label="แชร์"
              >
                <FiShare2 />
              </button>
            </div>
          </div>

          <h3 className="post-title">{post.title}</h3>
          <p className="description">{post.description}</p>

          {/* แท็ก */}
          {post.tags && post.tags.length > 0 && (
            <div className="post-tags">
              {post.tags.map((t, i) => (
                <span className="tag" key={`${t}-${i}`}>
                  #{t}
                </span>
              ))}
            </div>
          )}
        </section>
      </main>
    </div>
  );
};

export default PostDetail;