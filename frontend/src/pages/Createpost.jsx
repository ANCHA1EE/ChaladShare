import React, { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useNotification } from "../component/Notification";
import axios from "axios";
import Sidebar from "./Sidebar";
import Footer from "../component/Footer";
import "../component/Createpost.css";

const MAX_FILE_MB = 30; // จำกัดขนาดไฟล์ 10MB
const ACCEPTED_MIME = ["application/pdf"];

const MAX_COVER_MB = 5;
const ACCEPTED_COVER_MIME = ["image/jpeg", "image/png"];

function parseTags(input) {
  return input
    .split(/[,\s]+/g)
    .map((t) => t.trim().replace(/^#/, "").toLowerCase())
    .filter(Boolean);
}

const CreatePost = () => {
  const [formData, setForm] = useState({
    title: "",
    description: "",
    tags: "",
    visibility: "public",
    file: null,
    cover: null,
  });

  const [isLoading, setIsLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState("");

  // ✅ เพิ่ม: error รายช่อง (สำหรับ * ที่ห้ามว่าง)
  const [fieldErrors, setFieldErrors] = useState({
    title: "",
    cover: "",
    file: "",
    tags: "",
  });

  const coverInputRef = useRef(null);
  const fileInputRef = useRef(null);
  const navTimerRef = useRef(null);
  const [coverPreviewUrl, setCoverPreviewUrl] = useState("");
  const navigate = useNavigate();
  const { success: notifySuccess, error: notifyError } = useNotification();

  useEffect(() => {
    return () => {
      if (coverPreviewUrl) URL.revokeObjectURL(coverPreviewUrl);
    };
  }, [coverPreviewUrl]);

  useEffect(() => {
    return () => {
      if (navTimerRef.current) clearTimeout(navTimerRef.current);
    };
  }, []);

  // ✅ เพิ่ม: validate เฉพาะช่องที่มี * (จำเป็นต้องมี)
  const validateForm = () => {
    const next = { title: "", cover: "", file: "", tags: "" };

    if (!formData.title || formData.title.trim().length === 0) {
      next.title = "กรุณากรอกหัวข้อ";
    } if (!formData.cover) next.cover = "กรุณาอัปโหลดรูปหน้าปก";
    if (!formData.file) next.file = "กรุณาอัปโหลดไฟล์ .pdf";
    if (parseTags(formData.tags).length === 0) next.tags = "กรุณากรอกแท็กอย่างน้อย 1 แท็ก";

    setFieldErrors(next);
    return !Object.values(next).some(Boolean);
  };

  const handleChange = (e) => {
    const { name, value } = e.target;

    setForm({
      ...formData,
      [name]: value,
    });

    // ✅ เพิ่ม: ล้าง error ของช่องที่กำลังแก้
    if (name === "title" || name === "tags") {
      setFieldErrors((prev) => ({ ...prev, [name]: "" }));
    }
  };

  const handleCoverChange = (e) => {
    setErrorMsg("");
    const f = e.target.files && e.target.files[0];
    if (!f) return;

    const sizeMB = f.size / (1024 * 1024);
    if (!ACCEPTED_COVER_MIME.includes(f.type)) {
      setErrorMsg("รองรับเฉพาะรูปภาพ .jpg .png สำหรับหน้าปก");
      e.target.value = "";
      return;
    }
    if (sizeMB > MAX_COVER_MB) {
      setErrorMsg(`ไฟล์หน้าปกใหญ่เกินไป (สูงสุด ${MAX_COVER_MB} MB)`);
      e.target.value = "";
      return;
    }

    // ทำ preview
    if (coverPreviewUrl) URL.revokeObjectURL(coverPreviewUrl);
    setCoverPreviewUrl(URL.createObjectURL(f));

    setForm({ ...formData, cover: f });
    // ✅ เพิ่ม: ล้าง error cover เมื่อเลือกสำเร็จ
    setFieldErrors((prev) => ({ ...prev, cover: "" }));
  };

  // เลือกไฟล์ + ตรวจสอบชนิด+ขนาด
  const handleFileChange = (e) => {
    setErrorMsg("");
    const f = e.target.files && e.target.files[0];
    if (!f) return;

    const sizeMB = f.size / (1024 * 1024);
    if (!ACCEPTED_MIME.includes(f.type)) {
      setErrorMsg("รองรับเฉพาะไฟล์ .pdf เท่านั้น");
      e.target.value = "";
      return;
    }
    if (sizeMB > MAX_FILE_MB) {
      setErrorMsg(`ไฟล์ใหญ่เกินไป (สูงสุด ${MAX_FILE_MB} MB)`);
      e.target.value = "";
      return;
    }
    setForm({ ...formData, file: f });
    // ✅ เพิ่ม: ล้าง error file เมื่อเลือกสำเร็จ
    setFieldErrors((prev) => ({ ...prev, file: "" }));
  };

  // เพิ่ม: ล้างหน้าปก
  const clearCover = (e) => {
    if (e) {
      e.preventDefault();
      e.stopPropagation();
    }
    if (coverPreviewUrl) URL.revokeObjectURL(coverPreviewUrl);
    setCoverPreviewUrl("");
    setForm((prev) => ({ ...prev, cover: null }));
    // ✅ เพิ่ม: เคลียร์ error (กันค้าง)
    setFieldErrors((prev) => ({ ...prev, cover: "" }));
    if (coverInputRef.current) coverInputRef.current.value = "";
  };

  // เพิ่ม: ล้างไฟล์ pdf
  const clearFile = (e) => {
    if (e) {
      e.preventDefault();
      e.stopPropagation();
    }
    setForm((prev) => ({ ...prev, file: null }));
    // ✅ เพิ่ม: เคลียร์ error (กันค้าง)
    setFieldErrors((prev) => ({ ...prev, file: "" }));
    if (fileInputRef.current) fileInputRef.current.value = "";
  };

  // โพสต์
  const handleSubmit = async (e) => {
    e.preventDefault();
    setErrorMsg("");

    // ✅ สำคัญ: ให้กดโพสต์ได้ แล้วค่อย validate เพื่อโชว์เตือนสีแดง
    const ok = validateForm();
    if (!ok) {
      // setErrorMsg("กรุณากรอกข้อมูลที่จำเป็นให้ครบถ้วน");
      return;
    }

    try {
      setIsLoading(true);
      let coverUrl = null;

      if (formData.cover) {
        const coverData = new FormData();
        coverData.append("file", formData.cover);

        const coverRes = await axios.post("/files/cover", coverData, {
          withCredentials: true,
          headers: { "Content-Type": "multipart/form-data" },
        });

        coverUrl = coverRes.data && coverRes.data.cover_url;
        if (!coverUrl) {
          throw new Error("ไม่พบ cover url จากการอัปโหลดหน้าปก");
        }
      }

      // อัปโหลดไฟล์ PDF
      const fileData = new FormData();
      fileData.append("file", formData.file);
      const uploadRes = await axios.post("/files/doc", fileData, {
        withCredentials: true,
        headers: { "Content-Type": "multipart/form-data" },
      });

      const documentId = uploadRes.data && uploadRes.data.document_id;
      if (!documentId) throw new Error("ไม่พบ document id จากการอัปโหลด");

      // สร้างโพสต์
      const postData = {
        post_title: formData.title.trim(),
        post_description: formData.description.trim(),
        post_visibility: formData.visibility,
        document_id: documentId,
        cover_url: coverUrl,
        tags: parseTags(formData.tags),
      };

      await axios.post("/posts", postData, { withCredentials: true });
      notifySuccess("โพสต์สำเร็จ ", 5000);
      if (navTimerRef.current) clearTimeout(navTimerRef.current);
      navTimerRef.current = setTimeout(() => {
        handleCancel();
      }, 800);
    } catch (err) {
      if (navTimerRef.current) clearTimeout(navTimerRef.current);

      if (err && err.response && err.response.status === 401) {
        notifyError("กรุณาเข้าสู่ระบบใหม่");
        return navigate("/", { replace: true });
      }

      console.error("Create post error:", err);

      const msg =
        (err.response && err.response.data && err.response.data.error) ||
        err.message ||
        "เกิดข้อผิดพลาดในการโพสต์";

      setErrorMsg(msg);
      notifyError(msg, 2500);
    } finally {
      setIsLoading(false);
    }
  };

  const handleCancel = () => {
    if (coverPreviewUrl) URL.revokeObjectURL(coverPreviewUrl);
    setCoverPreviewUrl("");

    setForm({
      title: "",
      description: "",
      tags: "",
      visibility: "public",
      file: null,
      cover: null,
    });
    setErrorMsg("");
    // ✅ เพิ่ม: reset error รายช่อง
    setFieldErrors({ title: "", cover: "", file: "", tags: "" });

    if (coverInputRef.current) coverInputRef.current.value = "";
    if (fileInputRef.current) fileInputRef.current.value = "";

    navigate("/home");
  };

  return (
    <div className="create-page">
      <Sidebar />
      <div className="create-post-container">
        <h2 className="create-title">สร้างโพสต์ใหม่</h2>

        <form className="create-form" onSubmit={handleSubmit}>
          {/* หัวข้อ + visibility */}
          <div className="form-group">
            <label>
              หัวข้อ<span className="required">*</span>
            </label>
            <div className="title-row">
              <input
                type="text"
                name="title"
                placeholder="พิมพ์หัวข้อของคุณ..."
                value={formData.title}
                onChange={handleChange}
                disabled={isLoading}
                // ✅ เพิ่ม: ขอบแดงเมื่อผิด
                style={{
                  border: fieldErrors.title ? "1px solid #ff4d4f" : undefined,
                }}
              />
              <select
                name="visibility"
                value={formData.visibility}
                onChange={handleChange}
                disabled={isLoading}
              >
                <option value="public">สาธารณะ</option>
                <option value="friends">เฉพาะเพื่อน</option>
              </select>
            </div>

            {/* ✅ เพิ่ม: ข้อความ error ใต้ช่องหัวข้อ */}
            {fieldErrors.title ? (
              <div style={{ color: "#ff4d4f", marginTop: 6 }}>{fieldErrors.title}</div>
            ) : null}
          </div>

          {/* แสดง error (คงไว้เหมือนเดิม/ถ้าบีมอยากโชว์) */}
          {errorMsg ? (
            <div style={{ color: "red", marginBottom: 12 }}>{errorMsg}</div>
          ) : null}

          {/* อัปโหลดหน้าปก */}
          <div className="form-group">
            <label>
              รูปหน้าปก<span className="required">*</span>
            </label>
            <div
              className="upload-box"
              // ✅ เพิ่ม: ขอบแดงเมื่อผิด
              style={{
                border: fieldErrors.cover ? "1px solid #ff4d4f" : undefined,
              }}
            >
              <input
                ref={coverInputRef}
                type="file"
                id="cover-upload"
                onChange={handleCoverChange}
                accept="image/*"
                disabled={isLoading}
              />

              {/* ปุ่มล้างหน้าปก */}
              {formData.cover ? (
                <button
                  type="button"
                  className="create-clear-btn"
                  onClick={clearCover}
                  disabled={isLoading}
                  aria-label="ล้างรูปหน้าปก"
                >
                  ✕
                </button>
              ) : null}

              <label htmlFor="cover-upload" className="upload-label">
                {formData.cover ? (
                  <>
                    {/* preview รูป */}
                    {coverPreviewUrl ? (
                      <img
                        src={coverPreviewUrl}
                        alt="cover preview"
                        className="create-cover-preview"
                      />
                    ) : null}
                    <span>{formData.cover.name}</span>
                  </>
                ) : (
                  <>
                    {/* ไอคอนเดิมยังอยู่ */}
                    <img
                      src="https://cdn-icons-png.flaticon.com/512/1829/1829586.png"
                      alt="cover"
                      className="upload-icon"
                    />
                    <p>เพิ่มรูปหน้าปก</p>
                  </>
                )}
              </label>
            </div>

            {/* ✅ เพิ่ม: ข้อความ error ใต้กล่องหน้าปก */}
            {fieldErrors.cover ? (
              <div style={{ color: "#ff4d4f", marginTop: 6 }}>{fieldErrors.cover}</div>
            ) : null}
          </div>

          {/* อัปโหลดไฟล์ */}
          <div className="form-group">
            <label>
              อัปโหลดไฟล์<span className="required">*</span>
            </label>
            <div
              className="upload-box"
              // ✅ เพิ่ม: ขอบแดงเมื่อผิด
              style={{
                border: fieldErrors.file ? "1px solid #ff4d4f" : undefined,
              }}
            >
              <input
                ref={fileInputRef}
                type="file"
                id="file-upload"
                onChange={handleFileChange}
                accept=".pdf,application/pdf"
                disabled={isLoading}
              />

              {/* ปุ่มล้างไฟล์ pdf */}
              {formData.file ? (
                <button
                  type="button"
                  className="create-clear-btn"
                  onClick={clearFile}
                  disabled={isLoading}
                  aria-label="ล้างไฟล์ PDF"
                >
                  ✕
                </button>
              ) : null}

              <label htmlFor="file-upload" className="upload-label">
                {formData.file ? (
                  <span>{formData.file.name}</span>
                ) : (
                  <>
                    {/* ไอคอนเดิมยังอยู่ */}
                    <img
                      src="https://cdn-icons-png.flaticon.com/512/864/864685.png"
                      alt="upload"
                      className="upload-icon"
                    />
                    <p>เพิ่มไฟล์</p>
                  </>
                )}
              </label>
            </div>

            {/* ✅ เพิ่ม: ข้อความ error ใต้กล่องไฟล์ */}
            {fieldErrors.file ? (
              <div style={{ color: "#ff4d4f", marginTop: 6 }}>{fieldErrors.file}</div>
            ) : null}
          </div>

          {/* คำอธิบาย (กลับมาเหมือนเดิม) */}
          <div className="form-group">
            <label>คำอธิบาย</label>
            <textarea
              name="description"
              placeholder="เพิ่มรายละเอียดเกี่ยวกับโพสต์ของคุณ..."
              value={formData.description}
              onChange={handleChange}
              disabled={isLoading}
            />
          </div>

          {/* แท็ก (กลับมาเหมือนเดิม) */}
          <div className="form-group">
            <label>
              แท็ก<span className="required">*</span>
            </label>
            <input
              type="text"
              name="tags"
              placeholder="เช่น #uml #se หรือ uml,se"
              value={formData.tags}
              onChange={handleChange}
              disabled={isLoading}
              // ✅ เพิ่ม: ขอบแดงเมื่อผิด
              style={{
                border: fieldErrors.tags ? "1px solid #ff4d4f" : undefined,
              }}
            />

            {/* ✅ เพิ่ม: ข้อความ error ใต้ช่องแท็ก */}
            {fieldErrors.tags ? (
              <div style={{ color: "#ff4d4f", marginTop: 6 }}>{fieldErrors.tags}</div>
            ) : null}
          </div>

          {/* ปุ่ม */}
          <div className="button-group">
            <button
              type="button"
              className="btn-cancel"
              onClick={handleCancel}
              disabled={isLoading}
            >
              ยกเลิก
            </button>
            <button
              type="submit"
              className="btn-submit"
              // ✅ สำคัญ: ไม่ disable จากฟอร์มแล้ว เพื่อให้กดแล้วโชว์เตือนได้
              disabled={isLoading}
            >
              {isLoading ? "กำลังโพสต์..." : "โพสต์"}
            </button>
          </div>
        </form>
      </div>
      <Footer />
    </div>
  );
};

export default CreatePost;