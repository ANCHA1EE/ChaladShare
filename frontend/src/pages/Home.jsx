// หน้า Home.jsx (ทำ prefix แล้ว)

import React, { useMemo, useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { IoSearch } from "react-icons/io5";
import axios from "axios";
import Sidebar from "./Sidebar";
import PostCard from "../component/Postcard";
import RankingCard from "../component/RankingCard";
import Avatar from "../assets/default.png";
import PdfPlaceholder from "../assets/default.png";
import Footer from "../component/Footer";

import "../component/Home.css";

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

const mapToCardPost = (p) => {
  const coverRaw = p.cover_url || p.post_cover_url || "";
  const coverImg = coverRaw ? toAbsUrl(coverRaw) : PdfPlaceholder;

  const rawUrl = p.file_url || p.document_url || "";
  const isPdf = /\.pdf$/i.test(rawUrl || "");

  const avatarRaw = p.avatar_url || p.author_img || "";
  const authorImg = avatarRaw ? toAbsUrl(avatarRaw) : Avatar;

  const tagsRaw = p.tags ?? "";
  const tags = Array.isArray(tagsRaw)
    ? tagsRaw.map((t) => (t.startsWith("#") ? t : `#${t}`)).join(" ")
    : typeof tagsRaw === "string"
      ? tagsRaw
          .split(",")
          .map((t) => t.trim())
          .filter(Boolean)
          .map((t) => (t.startsWith("#") ? t : `#${t}`))
          .join(" ")
      : "";

  return {
    id: p.post_id ?? p.PostID ?? p.id,
    post_id: p.post_id ?? p.PostID ?? p.id,

    img: coverImg,

    isPdf,
    document_url: isPdf && rawUrl ? toAbsUrl(rawUrl) : null,
    documentId: p.post_document_id ?? p.document_id,

    likes: p.like_count ?? p.likeCount ?? 0,
    like_count: p.like_count ?? p.likeCount ?? 0,
    is_liked: !!(p.is_liked ?? p.isLiked),
    is_saved: !!(p.is_saved ?? p.isSaved),

    title: p.post_title ?? p.title ?? p.Title ?? "",
    tags,

    authorName:
      p.author_name ??
      p.authorName ??
      p.username ??
      p.user_username ??
      "ไม่ระบุ",
    authorImg,
  };
};

const Home = () => {
  //โพสต์ยอดนิยม
  const [popularPosts, setPopularPosts] = useState([]);
  const [loadingPop, setLoadingPop] = useState(true);
  const [popErr, setPopErr] = useState("");

  // แนะนำ
  const [recommendedPosts, setRecommendedPosts] = useState([]);
  const [loadingRec, setLoadingRec] = useState(true);
  const [recErr, setRecErr] = useState("");

  // โพสต์ทั้งหมด
  const [allPosts, setAllPosts] = useState([]);
  const [loadingAll, setLoadingAll] = useState(true);
  const [allErr, setAllErr] = useState("");

  const navigate = useNavigate();

  // ค้นหาโพสต์
  const [search, setSearch] = useState("");
  const [searchPage, setSearchPage] = useState(1);
  const searchSize = 20;

  const [searchPosts, setSearchPosts] = useState([]);
  const [searchTotal, setSearchTotal] = useState(0);
  const [loadingSearch, setLoadingSearch] = useState(false);
  const [searchErr, setSearchErr] = useState("");
  const isSearchMode = search.trim().length > 0;
  const normalizeQuery = (s) => (s || "").trim().replace(/^#+/, "");

  const patchPostEverywhere = (postId, patch) => {
    const apply = (setter) =>
      setter((prev) =>
        prev.map((p) => (p.id === postId ? { ...p, ...patch } : p)),
      );

    apply(setAllPosts);
    apply(setSearchPosts);
    apply(setRecommendedPosts);
    apply(setPopularPosts);
  };

  // เรียงโพสต์ยอดนิยมจากไลก์
  const rankedPopular = useMemo(() => {
    return popularPosts
      .slice()
      .sort(
        (a, b) =>
          (b.like_count ?? b.likes ?? 0) - (a.like_count ?? a.likes ?? 0),
      );
  }, [popularPosts]);

  const fetchSearchPosts = useCallback(
    async (q = search, p = searchPage) => {
      const qq = normalizeQuery(q);

      if (!qq) {
        setSearchPosts([]);
        setSearchTotal(0);
        setSearchErr("");
        setLoadingSearch(false);
        return;
      }

      setLoadingSearch(true);
      setSearchErr("");

      try {
        const res = await axios.get("/posts/search", {
          params: { search: qq, page: p, size: searchSize },
          withCredentials: true,
        });

        const payload = res?.data?.data ?? res?.data ?? {};
        const itemsRaw = Array.isArray(payload?.items) ? payload.items : [];
        const totalRaw = Number.isFinite(payload?.total) ? payload.total : 0;

        const mapped = itemsRaw.map(mapToCardPost);
        setSearchPosts(mapped);
        setSearchTotal(totalRaw);
      } catch (e) {
        if (e?.response?.status === 401) {
          navigate("/", { replace: true });
          return;
        }
        setSearchErr(
          e?.response?.data?.error || e.message || "ค้นหาโพสต์ล้มเหลว",
        );
        setSearchPosts([]);
        setSearchTotal(0);
      } finally {
        setLoadingSearch(false);
      }
    },
    [navigate, search, searchPage],
  );

  // โหลดโพสต์ยอดเยี่ยม
  useEffect(() => {
    let cancelled = false;

    (async () => {
      try {
        setLoadingPop(true);
        setPopErr("");

        const res = await axios.get("/posts/popular", {
          params: { limit: 3 },
          withCredentials: true,
        });

        const rows = Array.isArray(res?.data?.data) ? res.data.data : [];
        const mapped = rows.map(mapToCardPost);

        if (!cancelled) setPopularPosts(mapped);
      } catch (e) {
        if (!cancelled) {
          if (e?.response?.status === 401) {
            navigate("/", { replace: true });
            return;
          }
          setPopErr(
            e?.response?.data?.error ||
              e.message ||
              "โหลดโพสต์ยอดเยี่ยมล้มเหลว",
          );
        }
      } finally {
        if (!cancelled) setLoadingPop(false);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [navigate]);

  const recCards = recommendedPosts;
  // โหลดแนะนำ
  useEffect(() => {
    let cancelled = false;

    (async () => {
      try {
        setLoadingRec(true);
        setRecErr("");

        const res = await axios.get("/recommend", {
          params: { limit: 3 },
          withCredentials: true,
        });

        const rows = Array.isArray(res?.data?.data) ? res.data.data : [];
        const mapped = rows.map(mapToCardPost);

        if (!cancelled) setRecommendedPosts(mapped);
      } catch (e) {
        if (!cancelled) {
          if (e?.response?.status === 401) {
            navigate("/", { replace: true });
            return;
          }
          setRecErr(
            e?.response?.data?.error || e.message || "โหลดแนะนำล้มเหลว",
          );
        }
      } finally {
        if (!cancelled) setLoadingRec(false);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [navigate]);

  // โหลดโพสต์ทั้งหมด
  useEffect(() => {
    let cancelled = false;

    (async () => {
      try {
        setLoadingAll(true);
        setAllErr("");

        const res = await axios.get("/posts", { withCredentials: true });

        const rows = Array.isArray(res?.data?.data)
          ? res.data.data
          : Array.isArray(res?.data)
            ? res.data
            : [];

        const mapped = rows.map(mapToCardPost);

        if (!cancelled) setAllPosts(mapped);
      } catch (e) {
        if (!cancelled) {
          if (e?.response?.status === 401) {
            navigate("/", { replace: true });
            return;
          }
          setAllErr(
            e?.response?.data?.error || e.message || "โหลดโพสต์ทั้งหมดล้มเหลว",
          );
        }
      } finally {
        if (!cancelled) setLoadingAll(false);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [navigate]);

  // ค้นหา
  useEffect(() => {
    const t = setTimeout(() => {
      setSearchPage(1);
      fetchSearchPosts(search, 1);
    }, 300);

    return () => clearTimeout(t);
  }, [search, fetchSearchPosts]);

  // เปลี่ยนหน้า search
  useEffect(() => {
    if (!search.trim()) return;
    fetchSearchPosts(search, searchPage);
  }, [searchPage, search, fetchSearchPosts]);

  const goToPostDetail = (post) => {
    if (post?.id) navigate(`/posts/${post.id}`);
  };

  return (
    <div className="home-page">
      <div className="home-container">
        {/* Sidebar */}
        <Sidebar />

        {/* เนื้อหาหลัก */}
        <div className="home">
          {/* Search bar */}
          <div className="search-bar">
            <input
              type="text"
              placeholder="ค้นหาความสนใจของคุณ"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  setSearchPage(1);
                  fetchSearchPosts(search, 1);
                }
              }}
            />
            <IoSearch />
          </div>

          {isSearchMode && (
            <>
              {loadingSearch && <div>กำลังค้นหา...</div>}
              {searchErr && <div style={{ color: "#b00020" }}>{searchErr}</div>}

              {!loadingSearch && !searchErr && (
                <div className="card-list">
                  {searchPosts.length === 0 ? (
                    <div>ไม่พบโพสต์ที่ตรงกับคำค้น</div>
                  ) : (
                    searchPosts.map((post) => (
                      <div
                        key={post.id}
                        onClick={() => goToPostDetail(post)}
                        style={{ cursor: "pointer" }}
                      >
                        <PostCard
                          post={post}
                          onToggleLike={(id, is_liked, like_count) =>
                            patchPostEverywhere(id, {
                              is_liked,
                              like_count,
                              likes: like_count,
                            })
                          }
                          onToggleSave={(id, is_saved) =>
                            patchPostEverywhere(id, { is_saved })
                          }
                        />
                      </div>
                    ))
                  )}
                </div>
              )}

              {searchTotal > searchSize && (
                <div className="search-pagination">
                  <button
                    disabled={searchPage <= 1 || loadingSearch}
                    onClick={() => setSearchPage((p) => p - 1)}
                  >
                    ก่อนหน้า
                  </button>

                  <span>
                    {searchPage} / {Math.ceil(searchTotal / searchSize)}
                  </span>

                  <button
                    disabled={
                      searchPage >= Math.ceil(searchTotal / searchSize) ||
                      loadingSearch
                    }
                    onClick={() => setSearchPage((p) => p + 1)}
                  >
                    ถัดไป
                  </button>
                </div>
              )}
            </>
          )}

          {!isSearchMode && (
            <>
              {/* โพสต์ยอดนิยม */}
              <h3>โพสต์สรุปยอดเยี่ยม</h3>
              {loadingPop && <div>กำลังโหลดโพสต์ยอดเยี่ยม...</div>}
              {popErr && <div style={{ color: "#b00020" }}>{popErr}</div>}

              <div className="card-list">
                {!loadingPop && !popErr && rankedPopular.length === 0 && (
                  <div>ยังไม่มีโพสต์ยอดเยี่ยม</div>
                )}

                {!loadingPop &&
                  !popErr &&
                  rankedPopular.map((post, index) => (
                    <div
                      key={post.id ?? index}
                      onClick={() => goToPostDetail(post)}
                      style={{ cursor: "pointer" }}
                    >
                      <RankingCard
                        post={post}
                        rank={index + 1}
                        onToggleLike={(id, is_liked, like_count) =>
                          patchPostEverywhere(id, {
                            is_liked,
                            like_count,
                            likes: like_count,
                          })
                        }
                        onToggleSave={(id, is_saved) =>
                          patchPostEverywhere(id, { is_saved })
                        }
                      />
                    </div>
                  ))}
              </div>

              {/* แนะนำสรุปน่าอ่าน */}
              <h3>แนะนำสรุปน่าอ่าน</h3>
              {loadingRec && <div>กำลังโหลดแนะนำ...</div>}
              {recErr && <div style={{ color: "#b00020" }}>{recErr}</div>}
              <div className="card-list">
                {!loadingRec &&
                  !recErr &&
                  recCards.map((post) => (
                    <div
                      key={post.id ?? post.post_id}
                      onClick={() => goToPostDetail(post)}
                      style={{ cursor: "pointer" }}
                    >
                      <PostCard
                        post={post}
                        onToggleLike={(id, is_liked, like_count) =>
                          patchPostEverywhere(id, {
                            is_liked,
                            like_count,
                            likes: like_count,
                          })
                        }
                        onToggleSave={(id, is_saved) =>
                          patchPostEverywhere(id, { is_saved })
                        }
                      />
                    </div>
                  ))}
              </div>

              {/* โพสต์ทั้งหมด*/}
              <h3>โพสต์ทั้งหมด</h3>
              {loadingAll && <div>กำลังโหลดโพสต์ทั้งหมด...</div>}
              {allErr && <div style={{ color: "#b00020" }}>{allErr}</div>}
              <div className="card-list">
                {!loadingAll &&
                  !allErr &&
                  allPosts.map((post) => (
                    <div
                      key={post.id}
                      onClick={() => goToPostDetail(post)}
                      style={{ cursor: "pointer" }}
                    >
                      <PostCard
                        post={post}
                        onToggleLike={(id, is_liked, like_count) =>
                          patchPostEverywhere(id, {
                            is_liked,
                            like_count,
                            likes: like_count,
                          })
                        }
                        onToggleSave={(id, is_saved) =>
                          patchPostEverywhere(id, { is_saved })
                        }
                      />
                    </div>
                  ))}
              </div>
            </>
          )}
        </div>
      </div>
      <Footer />
    </div>
  );
};

export default Home;
