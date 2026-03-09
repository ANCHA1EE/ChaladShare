import React, { useState } from "react";
import { AiFillHeart, AiOutlineHeart } from "react-icons/ai";
import { BsBookmark, BsBookmarkFill } from "react-icons/bs";
import { FiShare2 } from "react-icons/fi";
import axios from "axios";
import award from "../icon/award.png";
import medal from "../icon/medal.png";
import trophy from "../icon/trophy.png";

const RankingCard = ({ post, rank, onToggleLike, onToggleSave }) => {
  const badge = rank === 1 ? award : rank === 2 ? trophy : medal;

  const [liked, setLiked] = useState(!!post.is_liked);
  const [saved, setSaved] = useState(!!post.is_saved);
  const [likes, setLikes] = useState(post.like_count ?? post.likes ?? 0);
  const [toast, setToast] = useState("");

  const toggleLike = async (e) => {
    e.stopPropagation();
    try {
      const res = await axios.post(
        `/posts/${post.id}/like`,
        {},
        { withCredentials: true },
      );
      const { is_liked, like_count } = res.data.data;

      setLiked(!!is_liked);
      setLikes(like_count ?? 0);

      onToggleLike?.(post.id, !!is_liked, like_count ?? 0);
    } catch (err) {
      console.error("toggle like failed", err);
    }
  };

  const handleSave = async (e) => {
    e.stopPropagation();
    try {
      const res = await axios.post(
        `/posts/${post.id}/save`,
        {},
        { withCredentials: true },
      );
      const { is_saved } = res.data.data;

      setSaved(!!is_saved);
      onToggleSave?.(post.id, !!is_saved);

      setToast(!!is_saved ? "✔️ บันทึกแล้ว" : "❌ ยกเลิกบันทึก");
      setTimeout(() => setToast(""), 2000);
    } catch (err) {
      console.error("toggle save failed", err);
    }
  };

  const sharePost = async (e) => {
    e.stopPropagation();
    const url = window.location.origin + `/posts/${post.id}`; // แก้ url ให้ตรง
    try {
      if (navigator.share) await navigator.share({ title: post.title, url });
      else {
        await navigator.clipboard.writeText(url);
        setToast("📋 คัดลอกลิงก์แล้ว");
        setTimeout(() => setToast(""), 2000);
      }
    } catch {}
  };

  return (
    <div className="card ranking-card">
      {rank <= 3 && (
        <img src={badge} alt={`อันดับ ${rank}`} className="rank-badge" />
      )}

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
              <AiFillHeart style={{ color: "red", fontSize: 20 }} />
            ) : (
              <AiOutlineHeart style={{ color: "black", fontSize: 20 }} />
            )}
            <span>{likes}</span>
          </span>

          <div className="action-right">
            <button
              className={`icon-btn ${saved ? "active" : ""}`}
              onClick={handleSave}
              title={saved ? "ยกเลิกบันทึก" : "บันทึก"}
            >
              {saved ? <BsBookmarkFill /> : <BsBookmark />}
            </button>

            <button className="icon-btn" onClick={sharePost} title="แชร์">
              <FiShare2 />
            </button>
          </div>
        </div>

        <h4>{post.title}</h4>
        <p>{post.tags}</p>
      </div>

      {toast && <div className="mini-toast">{toast}</div>}
    </div>
  );
};

export default RankingCard;
