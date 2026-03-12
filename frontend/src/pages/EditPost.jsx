// src/pages/EditPost.jsx
import React, { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import axios from "axios";

import Sidebar from "./Sidebar";
import "../component/Profile.css";

// const API_HOST = "http://localhost:8080";
const API_URL =
  process.env.REACT_APP_API_URL || "http://localhost:8080/api/v1";

const FILE_BASE_URL = API_URL.replace(/\/api\/v1\/?$/, "");

const toAbsUrl = (p) => {
  if (!p) return "";
  if (p.startsWith("http")) return p;
  const clean = p.replace(/^\./, "");
  return `${FILE_BASE_URL}${clean.startsWith("/") ? clean : `/${clean}`}`;
};
// ใช้ logic เดียวกับหน้า CreatePost
function parseTags(input) {
  return input
    .split(/[,\s]+/g)
    .map((t) => t.trim().replace(/^#/, "").toLowerCase())
    .filter(Boolean);
}

const EditPost = () => {
  const { id } = useParams();
  const navigate = useNavigate();

  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [successMsg, setSuccessMsg] = useState(""); // ข้อความเขียว

  const [postData, setPostData] = useState({
    title: "",
    description: "",
    tagsInput: "",
    visibility: "public",
    coverUrl: "",
    fileName: "",
  });

  // โหลดข้อมูลโพสต์มาเติมฟอร์ม + รูปปก + ชื่อไฟล์
  useEffect(() => {
    let cancelled = false;

    const fetchPost = async () => {
      try {
        setLoading(true);
        setError("");
        setSuccessMsg("");

        const res = await axios.get(`/posts/${id}`, {
          withCredentials: true,
        });

        // เผื่อหลายรูปแบบ response
        const raw =
          res?.data?.data?.post ||
          res?.data?.data ||
          res?.data?.post ||
          res?.data ||
          {};

        // 1) หัวข้อ / คำอธิบาย / visibility
        const title = raw.post_title || raw.title || "";
        const description = raw.post_description || raw.description || "";
        const visibility = raw.post_visibility || "public";

        // 2) tags → แสดงเป็น string ให้ user เห็นแบบ #ai #study
        let tagsInput = "";
        if (Array.isArray(raw.tags)) {
          tagsInput = raw.tags
            .map((t) => (typeof t === "string" ? t : ""))
            .filter(Boolean)
            .map((t) => (t.startsWith("#") ? t : `#${t}`))
            .join(" ");
        } else if (typeof raw.tags === "string") {
          tagsInput = raw.tags;
        }

        // 3) รูปปก
        const coverRaw = raw.cover_url || raw.cover || raw.coverPath || "";
        const coverUrl = coverRaw ? toAbsUrl(coverRaw) : "";

        // 4) ชื่อไฟล์
        const fileRaw =
          raw.file_url || raw.document_url || raw.document_path || raw.file || "";
        const fileNameDirect =
          raw.document_name ||
          raw.documentName ||
          raw.document_original_name ||
          "";

        const fileName = fileNameDirect
          ? fileNameDirect
          : fileRaw
          ? fileRaw.split("/").pop()
          : "";

        if (!cancelled) {
          setPostData({
            title,
            description,
            tagsInput,
            visibility,
            coverUrl,
            fileName,
          });
        }
      } catch (e) {
        if (!cancelled) {
          setError(
            e?.response?.data?.error || e.message || "โหลดโพสต์ไม่สำเร็จ"
          );
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    };

    fetchPost();
    return () => {
      cancelled = true;
    };
  }, [id]);

  const handleChangeField = (field) => (e) => {
    const value = e.target.value;
    setPostData((prev) => ({ ...prev, [field]: value }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setSuccessMsg("");

    // Validation (แก้เฉพาะ "คำอธิบาย" ให้เว้นว่างได้)
    if (!postData.title.trim()) {
      setError("กรุณากรอกหัวข้อสรุป");
      return;
    }

    const tags = parseTags(postData.tagsInput);
    if (tags.length === 0) {
      setError("กรุณากรอกอย่างน้อย 1 แท็ก");
      return;
    }

    try {
      setSaving(true);

      const payload = {
        post_title: postData.title.trim(),
        // คำอธิบายเว้นว่างได้: ส่งเป็นสตริง (trim เพื่อความสะอาด แต่ไม่บังคับว่าห้ามว่าง)
        post_description: (postData.description || "").trim(),
        post_visibility: postData.visibility,
        tags: tags,
      };

      await axios.put(`/posts/${id}`, payload, { withCredentials: true });

      // แสดงข้อความเขียวในหน้า EditPost
      setSuccessMsg("บันทึกการแก้ไขเรียบร้อยแล้ว");

      // รอให้ผู้ใช้เห็นข้อความสักครู่ แล้วค่อยไปหน้ารายละเอียดโพสต์
      setTimeout(() => {
        navigate(`/posts/${id}`, { replace: true });
      }, 800);
    } catch (e) {
      setSuccessMsg("");
      setError(e?.response?.data?.error || e.message || "บันทึกไม่สำเร็จ");
    } finally {
      setSaving(false);
    }
  };

  const handleCancel = () => {
    // ย้อนกลับหน้าที่มา (ส่วนใหญ่จะเป็นโปรไฟล์)
    navigate(-1);
  };

  if (loading) {
    return (
      <div className="profile-page">
        <div className="profile-container">
          <Sidebar />
          <main className="profile-content">
            <div className="profile-shell">
              <p className="profile-msg">กำลังโหลดโพสต์...</p>
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
            <section className="edit-card">
              <h2 style={{ marginBottom: 12 }}>แก้ไขโพสต์ของฉัน</h2>

              {/* แถว รูปปก + ไฟล์สรุป (อ่านอย่างเดียว) */}
              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "260px 1fr",
                  gap: "20px",
                  marginBottom: "18px",
                }}
              >
                {/* รูปปก */}
                <div
                  style={{
                    border: "1px solid #dbe3ee",
                    borderRadius: 12,
                    padding: 10,
                    textAlign: "center",
                  }}
                >
                  <p
                    style={{
                      fontSize: 14,
                      fontWeight: 600,
                      marginBottom: 8,
                    }}
                  >
                    รูปปก (เปลี่ยนไม่ได้)
                  </p>
                  {postData.coverUrl ? (
                    <img
                      src={postData.coverUrl}
                      alt="cover"
                      style={{
                        width: "100%",
                        height: 150,
                        objectFit: "cover",
                        borderRadius: 8,
                      }}
                    />
                  ) : (
                    <div
                      style={{
                        width: "100%",
                        height: 150,
                        borderRadius: 8,
                        background: "#eef2f6",
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                        fontSize: 13,
                        color: "#64748b",
                      }}
                    >
                      ไม่มีรูปปก
                    </div>
                  )}
                  <p
                    style={{
                      marginTop: 10,
                      fontSize: 12,
                      color: "#6b7280",
                    }}
                  >
                    * ไม่สามารถเปลี่ยนรูปปกจากหน้านี้ได้
                  </p>
                </div>

                {/* ไฟล์สรุป */}
                <div
                  style={{
                    border: "1px solid #dbe3ee",
                    borderRadius: 12,
                    padding: 10,
                  }}
                >
                  <p
                    style={{
                      fontSize: 14,
                      fontWeight: 600,
                      marginBottom: 6,
                    }}
                  >
                    ไฟล์สรุป (เปลี่ยนไม่ได้)
                  </p>
                  <div
                    style={{
                      fontSize: 13,
                      color: "#1f2933",
                      padding: "8px 10px",
                      borderRadius: 8,
                      background: "#f9fafb",
                    }}
                  >
                    {postData.fileName || "ไม่พบชื่อไฟล์"}
                  </div>
                  <p
                    style={{
                      marginTop: 10,
                      fontSize: 12,
                      color: "#6b7280",
                    }}
                  >
                    * ไม่สามารถเปลี่ยนไฟล์สรุปจากหน้านี้ได้
                  </p>
                </div>
              </div>

              {/* ฟอร์มแก้ไขข้อความ */}
              <form onSubmit={handleSubmit} className="edit-form-col">
                <div className="edit-field">
                  <label>หัวข้อสรุป</label>
                  <input
                    type="text"
                    value={postData.title}
                    onChange={handleChangeField("title")}
                    placeholder="เช่น AI คืออะไร Part 3 ฉบับทดลองแก้ไข"
                  />
                </div>

                <div className="edit-field">
                  <label>คำอธิบายสรุป</label>
                  <textarea
                    rows={4}
                    value={postData.description}
                    onChange={handleChangeField("description")}
                    placeholder="อธิบายสั้น ๆ ว่าโพสต์นี้เกี่ยวกับอะไร (เว้นว่างได้)"
                  />
                </div>

                <div className="edit-field">
                  <label>
                    แท็ก (คั่นด้วยช่องว่าง เช่น #ai #study #note หรือ ai,study)
                  </label>
                  <input
                    type="text"
                    value={postData.tagsInput}
                    onChange={handleChangeField("tagsInput")}
                    placeholder="#ai #study #note"
                  />
                </div>

                {/* Success / Error message ที่ตำแหน่งเดียวกัน */}
                <div style={{ height: "24px", marginTop: "4px" }}>
                  {successMsg ? (
                    <p
                      style={{
                        color: "#16a34a",
                        margin: 0,
                        fontSize: "14px",
                        fontWeight: 500,
                      }}
                    >
                      {successMsg}
                    </p>
                  ) : error ? (
                    <p className="edit-error" style={{ margin: 0 }}>
                      {error}
                    </p>
                  ) : null}
                </div>

                <div className="edit-actions" style={{ marginTop: 16 }}>
                  <button
                    type="button"
                    className="btn-cancel"
                    onClick={handleCancel}
                    disabled={saving}
                  >
                    ยกเลิก
                  </button>
                  <button type="submit" className="btn-save" disabled={saving}>
                    {saving ? "กำลังบันทึก…" : "บันทึกการแก้ไข"}
                  </button>
                </div>
              </form>
            </section>
          </div>
        </main>
      </div>
    </div>
  );
};

export default EditPost;