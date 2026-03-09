import React, { useEffect, useState } from "react";
import { AiFillHeart, AiOutlineHeart } from "react-icons/ai";
import { BsBookmark, BsBookmarkFill } from "react-icons/bs";
import { FiShare2 } from "react-icons/fi";
import axios from "axios";

const PostCard = ({ post, onToggleSave, onToggleLike }) => {
  const postId = post.id ?? post.post_id;

  const [liked, setLiked] = useState(!!post.is_liked);
  const [likes, setLikes] = useState(
    typeof post.like_count === "number" ? post.like_count : post.likes || 0,
  );
  const [saved, setSaved] = useState(!!post.is_saved);
  const [toast, setToast] = useState("");

  useEffect(() => {
    setLiked(!!post.is_liked);
    setSaved(!!post.is_saved);

    const lc =
      typeof post.like_count === "number"
        ? post.like_count
        : typeof post.likes === "number"
          ? post.likes
          : 0;

    setLikes(lc);
  }, [postId, post.is_liked, post.is_saved, post.like_count, post.likes]);

  const toggleLike = async (e) => {
    e.stopPropagation();
    try {
      const res = await axios.post(`/posts/${postId}/like`, {}, { withCredentials: true });
      const { is_liked, like_count } = res.data.data;

      setLiked(is_liked);
      setLikes(like_count);

      onToggleLike?.(postId, is_liked, like_count);
    } catch (err) {
      console.error("toggle like error:", err);
      setToast("⚠️ กดถูกใจไม่สำเร็จ");
      setTimeout(() => setToast(""), 2500);
    }
  };

  const handleSave = async (e) => {
    e.stopPropagation();
    try {
      const res = await axios.post(`/posts/${postId}/save`, {}, { withCredentials: true });
      const { is_saved, save_count } = res.data.data;

      setSaved(is_saved);
      setToast(is_saved ? "✔️ บันทึกรายการแล้ว" : "❌ ยกเลิกการบันทึก");
      setTimeout(() => setToast(""), 2500);

      onToggleSave?.(postId, is_saved, save_count);
    } catch (err) {
      console.error("toggle save error:", err);
      setToast("⚠️ บันทึกโพสต์ไม่สำเร็จ");
      setTimeout(() => setToast(""), 2500);
    }
  };

  const sharePost = async (e) => {
    e.stopPropagation();
    const url =
      window.location.origin + "/post/" + encodeURIComponent(post.post_id);
    try {
      if (navigator.share) {
        await navigator.share({
          title: post.title,
          text: "ดูสรุปนี้บน ChaladShare",
          url,
        });
      } else {
        await navigator.clipboard.writeText(url);
        setToast("📋 คัดลอกลิงก์แล้ว");
        setTimeout(() => setToast(""), 3000);
      }
    } catch {}
  };

  return (
    <div className="card">
      <div className="card-header">
        <img src={post.authorImg} alt="author" className="author-img" />
        <span>{post.authorName}</span>
      </div>

      <img src={post.img} alt="summary" className="card-image" />

      <div className="card-body">
        <div className="actions-row" onClick={(e) => e.stopPropagation()}>
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
            {likes}
          </span>

          <div className="action-right">
            <button
              className={`icon-btn ${saved ? "active" : ""}`}
              onClick={handleSave}
              aria-label="save"
              title={saved ? "ยกเลิกบันทึก" : "บันทึก"}
            >
              {saved ? <BsBookmarkFill /> : <BsBookmark />}
            </button>

            <button
              className="icon-btn"
              onClick={sharePost}
              aria-label="share"
              title="แชร์"
            >
              <FiShare2 />
            </button>
          </div>
        </div>

        <h4>{post.title}</h4>
        <p>{post.tags}</p>
      </div>

      {/* Toast แยกออกจาก flow การ hover */}
      <div className="toast-container">
        {toast && <div className="mini-toast">{toast}</div>}
      </div>
    </div>
  );
};

export default PostCard;