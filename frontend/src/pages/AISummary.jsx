import React, { useEffect, useRef, useState } from "react";
import Sidebar from "./Sidebar";
import Footer from "../component/Footer";
import { MdLocalPrintshop } from "react-icons/md";
import { LuRefreshCcw } from "react-icons/lu";
import { PDFDocument } from "pdf-lib";

import "../component/AISummary.css";

const UploadIcon = () => (
    <svg width="48" height="48" viewBox="0 0 64 64" aria-hidden="true">
        <rect x="10" y="14" width="28" height="22" rx="3" fill="none" stroke="#0b5394" strokeWidth="2" />
        <rect x="26" y="26" width="28" height="22" rx="3" fill="none" stroke="#0b5394" strokeWidth="2" />
        <circle cx="20" cy="22" r="3" fill="#0b5394" opacity="0.9" />
        <path d="M14 34l7-7 6 6 5-4 6 5" fill="none" stroke="#0b5394" strokeWidth="2" strokeLinejoin="round" />
        <path d="M30 44l7-7 6 6 5-4 6 5" fill="none" stroke="#0b5394" strokeWidth="2" strokeLinejoin="round" />
    </svg>
);

const SparkleIcon = () => (
    <svg width="28" height="28" viewBox="0 0 64 64" aria-hidden="true">
        <path d="M32 6l4.5 16.5L53 27l-16.5 4.5L32 48l-4.5-16.5L11 27l16.5-4.5L32 6z" fill="#6ec1ff" />
        <path d="M50 38l2.6 9.2L62 50l-9.4 2.8L50 62l-2.6-9.2L38 50l9.4-2.8L50 38z" fill="#ff7aa2" />
    </svg>
);

const ReuseIcon = (props) => (
    <LuRefreshCcw size={18} aria-hidden="true" {...props} />
);

const PrintIcon = (props) => (
    <MdLocalPrintshop size={18} aria-hidden="true" {...props} />
);

// ✅ เรียกแบบเดิม: ยิง ngrok ตรง
const API_URL = "https://unsmarting-kamari-arbored.ngrok-free.dev";

// ✅ คีย์สำหรับเก็บ state หน้านี้
const LS_KEY = "chaladshare_ai_summary_state_v2";

// ✅ จำกัดจำนวนหน้าสูงสุด
const MAX_PDF_PAGES = 30;

// ✅ อ่านจำนวนหน้า PDF ด้วย pdf-lib
const getPdfPageCount = async (file) => {
    const buffer = await file.arrayBuffer();
    const pdfDoc = await PDFDocument.load(buffer);
    return pdfDoc.getPageCount();
};

// phase: "idle" | "processing" | "done" | "error"
const AISummary = () => {
    const inputRef = useRef(null);
    const abortRef = useRef(null);
    const statusTimerRef = useRef(null);
    const noticeTimerRef = useRef(null);

    const [file, setFile] = useState(null); // File จริง (รีเฟรชแล้วกู้คืนไม่ได้)
    const [isLoading, setIsLoading] = useState(false);
    const [errorMsg, setErrorMsg] = useState("");
    const [summaryHtml, setSummaryHtml] = useState("");
    const [statusMsg, setStatusMsg] = useState("");
    const [phase, setPhase] = useState("idle");

    // ✅ เพิ่ม: เก็บชื่อไฟล์เดิมไว้โชว์ได้ (แต่ไม่ทำให้ hasFile=true หลังรีเฟรช)
    const [restoredFileName, setRestoredFileName] = useState("");

    // ✅ toast แจ้งเตือนแบบมินิมอล
    const [notice, setNotice] = useState({
        open: false,
        title: "",
        message: "",
        type: "warning", // warning | error | success | info
    });

    const openNotice = (title, message, type = "warning") => {
        if (noticeTimerRef.current) {
            clearTimeout(noticeTimerRef.current);
            noticeTimerRef.current = null;
        }

        setNotice({
            open: true,
            title,
            message,
            type,
        });

        noticeTimerRef.current = setTimeout(() => {
            setNotice({
                open: false,
                title: "",
                message: "",
                type: "warning",
            });
            noticeTimerRef.current = null;
        }, 5000);
    };

    const closeNotice = () => {
        if (noticeTimerRef.current) {
            clearTimeout(noticeTimerRef.current);
            noticeTimerRef.current = null;
        }

        setNotice({
            open: false,
            title: "",
            message: "",
            type: "warning",
        });
    };

    const resetToIdle = () => {
        // ยกเลิกงานถ้ามี
        if (abortRef.current) abortRef.current.abort();
        abortRef.current = null;

        if (statusTimerRef.current) {
            clearTimeout(statusTimerRef.current);
            statusTimerRef.current = null;
        }

        setFile(null);
        setIsLoading(false);
        setErrorMsg("");
        setStatusMsg("");
        setPhase("idle");

        // ไม่ลบ summaryHtml ตรงนี้ (เพื่อให้ค้างผลเดิมได้ตาม requirement)
    };

    const onPickFile = () => {
        // ✅ ถ้ากำลังประมวลผล แล้วผู้ใช้กดเปลี่ยนไฟล์ -> abort แล้วให้เลือกไฟล์ใหม่ได้เลย
        if (isLoading) {
            resetToIdle();
            setSummaryHtml(""); // เริ่มใหม่ = เคลียร์ผลเดิม (เพราะ user จะอัปไฟล์ใหม่)
            setRestoredFileName("");
            try {
                localStorage.removeItem(LS_KEY);
            } catch { }
            setStatusMsg("ยกเลิกการประมวลผลแล้ว กรุณาเลือกไฟล์ใหม่");
        }

        inputRef.current?.click();
    };

    // ---------------------------
    // ✅ Restore state เมื่อรีเฟรชหน้า
    // กติกา:
    // - ถ้าตอนนั้นกำลัง processing แล้วรีเฟรช => กลับไปเหมือนไม่อัปโหลด (ต้องอัปใหม่)
    // - ถ้า done แล้ว => ค้างผลสรุป + กดพิมพ์ได้
    // - หลังรีเฟรช: ฝั่งซ้ายต้องกลับไปเป็นปุ่มรับไฟล์ (hasFile=false)
    // ---------------------------
    useEffect(() => {
        try {
            const raw = localStorage.getItem(LS_KEY);
            if (!raw) return;

            const saved = JSON.parse(raw);
            if (!saved) return;

            // ถ้าเคยกำลังประมวลผล แล้วรีเฟรช → ให้กลับไปเริ่มใหม่
            if (saved.phase === "processing") {
                localStorage.removeItem(LS_KEY);
                setFile(null);
                setIsLoading(false);
                setErrorMsg("");
                setSummaryHtml("");
                setStatusMsg("");
                setPhase("idle");
                setRestoredFileName("");
                return;
            }

            // ถ้าสรุปเสร็จแล้ว → ค้างผลลัพธ์ (แต่ฝั่งซ้ายให้เหมือนไม่มีไฟล์)
            if (saved.phase === "done" && saved.summaryHtml) {
                setSummaryHtml(saved.summaryHtml || "");
                setErrorMsg(saved.errorMsg || "");
                setStatusMsg(saved.statusMsg || "โหลดผลสรุปเดิมแล้ว ✅ หากต้องการสรุปใหม่ กรุณาอัปโหลดไฟล์อีกครั้ง");
                setPhase("done");
                setIsLoading(false);

                // ✅ สำคัญ: อย่า setFile เป็น restored (เพื่อให้ hasFile=false และกลับไปปุ่มอัปโหลด)
                setFile(null);
                setRestoredFileName(saved.fileName || "");
                return;
            }

            // กรณีอื่น ๆ ล้างทิ้ง
            localStorage.removeItem(LS_KEY);
        } catch {
            // ignore
        }
    }, []);

    // ---------------------------
    // ✅ Persist state
    // - เก็บเฉพาะตอน done เท่านั้น (ค้างเฉพาะ “ผลลัพธ์”)
    // ---------------------------
    useEffect(() => {
        try {
            if (phase !== "done" || !summaryHtml) {
                localStorage.removeItem(LS_KEY);
                return;
            }

            localStorage.setItem(
                LS_KEY,
                JSON.stringify({
                    phase,
                    fileName: file?.name || restoredFileName || "",
                    summaryHtml: summaryHtml || "",
                    errorMsg: errorMsg || "",
                    statusMsg: statusMsg || "",
                    ts: Date.now(),
                })
            );
        } catch {
            // ignore
        }
    }, [phase, file, restoredFileName, summaryHtml, errorMsg, statusMsg]);

    // ✅ ให้ข้อความ "สรุปเสร็จแล้ว" แสดง 5 วินาทีแล้วหายเอง
    useEffect(() => {
        if (statusMsg !== "สรุปเสร็จแล้ว") return;

        if (statusTimerRef.current) {
            clearTimeout(statusTimerRef.current);
        }

        statusTimerRef.current = setTimeout(() => {
            setStatusMsg("");
            statusTimerRef.current = null;
        }, 5000);

        return () => {
            if (statusTimerRef.current) {
                clearTimeout(statusTimerRef.current);
            }
        };
    }, [statusMsg]);

    // ✅ cleanup timers ตอน component ถูกถอด
    useEffect(() => {
        return () => {
            if (statusTimerRef.current) {
                clearTimeout(statusTimerRef.current);
            }
            if (noticeTimerRef.current) {
                clearTimeout(noticeTimerRef.current);
            }
        };
    }, []);

    const uploadToAI = async (pdfFile) => {
        if (isLoading) return;

        // ยกเลิก request เก่าถ้ายังวิ่งอยู่
        if (abortRef.current) abortRef.current.abort();
        const controller = new AbortController();
        abortRef.current = controller;

        // เริ่มงานใหม่ = ไม่ให้ค้างของเก่า
        setIsLoading(true);
        setErrorMsg("");
        setSummaryHtml("");
        console.log("กำลังส่งไฟล์ไปที่ AI...");
        setPhase("processing");
        setRestoredFileName("");

        try {
            console.log("1) เริ่มส่งไฟล์ไปที่ API:", `${API_URL}/summarize`, "file=", pdfFile?.name);

            const formData = new FormData();
            formData.append("file", pdfFile);

            const res = await fetch(`${API_URL}/summarize`, {
                method: "POST",
                body: formData,
                signal: controller.signal,
                credentials: "omit",
                headers: {
                    Accept: "application/json",
                    "ngrok-skip-browser-warning": "true",
                },
            });

            console.log("2) API ตอบกลับแล้ว status =", res.status);
            console.log(`AI รับไฟล์แล้ว (HTTP ${res.status}) กำลังประมวลผล...`);

            const ct = (res.headers.get("content-type") || "").toLowerCase();
            if (!ct.includes("application/json")) {
                const text = await res.text();
                throw new Error(
                    res.ok
                        ? "Response ไม่ใช่ JSON (อาจโดน ngrok warning/HTML แทรก)"
                        : `HTTP ${res.status}: ${text.slice(0, 200)}`
                );
            }

            const data = await res.json();
            if (!res.ok) throw new Error(data?.error || `HTTP ${res.status}`);

            setSummaryHtml(data?.summary_html || "");
            setStatusMsg("สรุปเสร็จแล้ว");
            setPhase("done");
        } catch (err) {
            if (err?.name === "AbortError") return;
            setErrorMsg(err?.message || "เกิดข้อผิดพลาดในการเชื่อมต่อกับ Server");
            setStatusMsg("เกิดข้อผิดพลาด");
            setPhase("error");
            openNotice("เชื่อมต่อไม่สำเร็จ", err?.message || "เกิดข้อผิดพลาดในการเชื่อมต่อกับ Server", "error");
        } finally {
            if (abortRef.current === controller) setIsLoading(false);
        }
    };

    const onFileChange = async (e) => {
        const f = e.target.files?.[0];
        if (!f) return;

        const isPdf = f.type === "application/pdf" || f.name.toLowerCase().endsWith(".pdf");
        if (!isPdf) {
            openNotice("ไฟล์ไม่ถูกต้อง", "กรุณาเลือกไฟล์ PDF เท่านั้น", "warning");
            if (inputRef.current) inputRef.current.value = "";
            return;
        }

        try {
            const totalPages = await getPdfPageCount(f);

            if (totalPages > MAX_PDF_PAGES) {
                openNotice(
                    "เอกสารมีจำนวนหน้าเกินกำหนด",
                    `เอกสารนี้มี ${totalPages} หน้า โดยระบบกำหนดจำนวนหน้าไม่เกิน ${MAX_PDF_PAGES} หน้า`,
                    "warning"
                );
                if (inputRef.current) inputRef.current.value = "";
                return;
            }
        } catch (err) {
            console.error("getPdfPageCount error:", err);

            openNotice(
                "ไม่สามารถตรวจจำนวนหน้าได้",
                "ระบบจะลองอัปโหลดไฟล์นี้ต่อ แต่หากเป็นไฟล์ขนาดใหญ่ อาจใช้เวลาประมวลผลนาน",
                "warning"
            );

            setFile(f);
            setSummaryHtml("");
            setErrorMsg("");
            setStatusMsg("");
            setPhase("idle");
            setRestoredFileName("");

            if (inputRef.current) inputRef.current.value = "";
            uploadToAI(f);
            return;
        }

        // เลือกไฟล์ใหม่ = เริ่มใหม่
        setFile(f);
        setSummaryHtml("");
        setErrorMsg("");
        setStatusMsg("");
        setPhase("idle");
        setRestoredFileName("");

        if (inputRef.current) inputRef.current.value = "";
        uploadToAI(f);
    };

    const onClear = () => {
        if (abortRef.current) abortRef.current.abort();

        if (statusTimerRef.current) {
            clearTimeout(statusTimerRef.current);
            statusTimerRef.current = null;
        }

        setFile(null);
        setSummaryHtml("");
        setErrorMsg("");
        setStatusMsg("");
        setIsLoading(false);
        setPhase("idle");
        setRestoredFileName("");

        try {
            localStorage.removeItem(LS_KEY);
        } catch { }

        if (inputRef.current) inputRef.current.value = "";
    };

    const onResummarize = async () => {
        if (isLoading) return;
        if (!file) return; // หลังรีเฟรชไม่มีไฟล์จริง -> ไม่โชว์ข้อความ ไม่ยิง API
        await uploadToAI(file);
    };

    // ✅ พิมพ์: print ครั้งเดียว + บังคับสี/ไฮไลท์ออกตอนพิมพ์
    const onPrint = () => {
        if (!summaryHtml) return;

        const iframe = document.createElement("iframe");
        iframe.style.position = "fixed";
        iframe.style.right = "0";
        iframe.style.bottom = "0";
        iframe.style.width = "0";
        iframe.style.height = "0";
        iframe.style.border = "0";

        document.body.appendChild(iframe);

        const w = iframe.contentWindow;
        const doc = w.document;

        // ✅ กัน print ซ้ำ (แก้ปัญหา cancel 2 รอบ)
        let printed = false;

        doc.open();
        doc.write(`
      <!doctype html>
      <html>
        <head>
          <meta charset="utf-8" />
          <meta name="viewport" content="width=device-width, initial-scale=1" />
          <title>ChaladShare - Summary</title>
          <style>
            body { font-family: Arial, sans-serif; padding: 24px; }
            * { -webkit-print-color-adjust: exact !important; print-color-adjust: exact !important; }
            mark, span { -webkit-print-color-adjust: exact !important; print-color-adjust: exact !important; }
          </style>
        </head>
        <body>
          ${summaryHtml}
        </body>
      </html>
    `);
        doc.close();

        const doPrintOnce = () => {
            if (printed) return;
            printed = true;

            w.focus();
            w.print();

            setTimeout(() => {
                try {
                    document.body.removeChild(iframe);
                } catch { }
            }, 700);
        };

        // บาง browser ไม่ยิง onload แบบคงที่หลัง doc.write → ใช้ timeout ช่วย
        iframe.onload = doPrintOnce;
        setTimeout(doPrintOnce, 200);
    };

    const onUploadKeyDown = (e) => {
        if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            onPickFile();
        }
    };

    const hasFile = Boolean(file);

    return (
        <div className="profile-page ai-page">
            <div className="profile-container">
                <Sidebar />

                <main className="profile-content">
                    <div className="profile-shell">
                        <div className="ai-layout">
                            {/* LEFT */}
                            <aside className="ai-source-panel">
                                <div className="ai-left-title">AI ช่วยสรุป</div>
                                <div className="ai-left-sub">แหล่งข้อมูล</div>

                                <input
                                    ref={inputRef}
                                    type="file"
                                    accept=".pdf,application/pdf"
                                    className="ai-hidden"
                                    onChange={onFileChange}
                                />

                                {!hasFile ? (
                                    <div
                                        className="ai-upload"
                                        role="button"
                                        tabIndex={0}
                                        onClick={onPickFile}
                                        onKeyDown={onUploadKeyDown}
                                        aria-label="อัปโหลดไฟล์ PDF"
                                        title="คลิกเพื่ออัปโหลดไฟล์ PDF"
                                    >
                                        <div className="ai-upload-icon">
                                            <UploadIcon />
                                        </div>
                                        <div className="ai-upload-text">เพิ่มไฟล์ PDF</div>
                                    </div>
                                ) : (
                                    <div className="ai-file">
                                        <div className="ai-file-row">
                                            <div className="ai-file-name" title={file?.name}>
                                                {file?.name}
                                            </div>
                                        </div>

                                        <div className="ai-file-actions">
                                            <button className="ai-btn ai-btn-ghost" type="button" onClick={onPickFile}>
                                                เปลี่ยนไฟล์
                                            </button>
                                            <button className="ai-btn ai-btn-danger" type="button" onClick={onClear}>
                                                ลบไฟล์
                                            </button>
                                        </div>

                                        {isLoading && (
                                            <div style={{ marginTop: 10, fontWeight: 800, color: "#0b5394" }}>
                                                <div>กำลังสรุป...</div>
                                                <div style={{ fontWeight: 500, marginTop: 4 }}>(อาจใช้เวลาสักครู่)</div>
                                            </div>
                                        )}

                                        {statusMsg && (
                                            <div style={{ marginTop: 10, fontWeight: 700, color: "#0b5394" }}>
                                                {statusMsg}
                                            </div>
                                        )}

                                        {errorMsg && (
                                            <div style={{ marginTop: 10, fontWeight: 800, color: "#b42318" }}>
                                                {errorMsg}
                                            </div>
                                        )}
                                    </div>
                                )}

                                {/* ✅ คำแนะนำ (อยู่นอกกรอบขาว + แสดงเสมอ) */}
                                <div className="ai-tip">
                                    <div className="ai-tip-title">คำแนะนำ</div>
                                    <div className="ai-tip-text">
                                        AI อาจให้ผลการสรุปแตกต่างกันในแต่ละครั้ง หากผลลัพธ์ยังไม่ตรงกับความต้องการสามารถกด <b>สรุปใหม่</b> <br />
                                        เพื่อประมวลผลอีกครั้ง
                                    </div>
                                </div>
                            </aside>

                            {/* RIGHT */}
                            <section className="ai-output-panel">
                                {!summaryHtml && (
                                    <div className="ai-greet">
                                        <div className="ai-loading-center">
                                            <SparkleIcon />
                                        </div>

                                        <div className="ai-greet-text">
                                            <div className="ai-greet-title">สวัสดี, ฉันคือ AI ที่จะช่วยสรุปเนื้อหาของคุณ</div>
                                            <div className="ai-greet-sub">อัปโหลดแหล่งข้อมูลแล้วเริ่มช่วยเหลือได้เลย</div>

                                            {isLoading && (
                                                <div style={{ marginTop: 12, fontWeight: 800, color: "#0b5394" }}>
                                                    กำลังประมวลผล กรุณารอสักครู่...
                                                </div>
                                            )}
                                            {isLoading && (
                                                <div className="ai-loading-box">
                                                    <SparkleIcon />
                                                </div>
                                            )}

                                            {errorMsg && (
                                                <div style={{ marginTop: 12, fontWeight: 800, color: "#b42318" }}>
                                                    {errorMsg}
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                )}

                                {summaryHtml && (
                                    <div className="ai-output-wrap">
                                        <button
                                            type="button"
                                            className="ai-resummarize-btn"
                                            onClick={(e) => {
                                                e.preventDefault();
                                                onResummarize();
                                            }}
                                            disabled={isLoading}
                                            aria-label="สรุปใหม่จากไฟล์เดิม"
                                            title="สรุปใหม่"
                                        >
                                            <ReuseIcon />
                                        </button>

                                        <button
                                            type="button"
                                            className="ai-resummarize-btn"
                                            style={{ right: 74 }}
                                            onClick={(e) => {
                                                e.preventDefault();
                                                onPrint();
                                            }}
                                            disabled={!summaryHtml || isLoading}
                                            aria-label="พิมพ์ผลสรุป"
                                            title="พิมพ์"
                                        >
                                            <PrintIcon />
                                        </button>

                                        <div className="ai-summary-render" dangerouslySetInnerHTML={{ __html: summaryHtml }} />
                                    </div>
                                )}
                            </section>
                        </div>
                    </div>
                </main>
            </div>

            <Footer />

            {notice.open && (
                <div className="ai-toast">
                    <div className={`ai-toast-card ai-toast-${notice.type}`}>
                        <div className="ai-toast-title">{notice.title}</div>
                        <div className="ai-toast-text">{notice.message}</div>
                        <button
                            type="button"
                            className="ai-toast-close"
                            onClick={closeNotice}
                            aria-label="ปิดการแจ้งเตือน"
                            title="ปิด"
                        >
                            ×
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default AISummary;