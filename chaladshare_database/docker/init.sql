CREATE EXTENSION IF NOT EXISTS vector;

-- ตาราง users สำหรับ login/register 172.20.10.2
CREATE EXTENSION IF NOT EXISTS citext;
create table if not exists users (
    user_id         serial primary key,                     -- id auto increment
    email           citext unique not null,                 -- email ไม่ซ้ำ ห้ามว่าง
    username        varchar(50) unique not null,            -- username ไม่ซ้ำ ห้ามว่าง
    username_ci     varchar(50) generated always as 
                     (lower(username)) stored,              -- ทำ index คำเล็ก (case-insensitive)
    password_hash   varchar(255) not null,                  -- เก็บรหัสผ่านแบบ hash
    user_created_at timestamptz default now(),              -- เวลาสร้าง
    user_status     varchar(20) default 'active'            -- สถานะ เช่น active / inactive
);

-- สร้าง unique index สำหรับ username_ci กันซ้ำแบบ case-insensitive
create unique index if not exists users_username_ci_uq on users(username_ci);

-------------------------------------------------------------------------------
-- ตารางโปรไฟล์ผู้ใช้
create table if not exists user_profiles (
    profile_user_id integer primary key
        references users(user_id) on delete cascade,
    avatar_url      varchar(255),
    avatar_storage  varchar(255),
    bio             varchar(150),
    created_at      timestamptz default now(),
    updated_at      timestamptz default now()
);

-- ฟังก์ชันสำหรับอัปเดต updated_at
create or replace function set_updated_at()
returns trigger as $$
begin
    new.updated_at := now();
    return new;
end;
$$ language plpgsql;

-- ทริกเกอร์สำหรับ user_profiles
drop trigger if exists trg_user_profiles_updated_at on user_profiles;
create trigger trg_user_profiles_updated_at
before update on user_profiles
for each row
execute function set_updated_at();

-- ตารางเก็บ session การล็อกอิน (ใช้ refresh token)
create table if not exists auth_sessions (
    session_id          serial primary key,
    session_user_id     integer NOT NULL references users(user_id) on delete cascade,
    refresh_token_hash  varchar(255) not null,          -- เก็บ hash
    session_expires_at  timestamptz not null,
    revoked_at          timestamptz,
    created_at          timestamptz default now(),
    last_used_at        timestamptz,
    replaced_by_session_id integer references auth_sessions(session_id)
);

-- session ที่ยัง active ของ user
create index if not exists ix_auth_sessions_user_active
  on auth_sessions(session_user_id)
  where revoked_at is null;

CREATE UNIQUE INDEX IF NOT EXISTS ux_auth_sessions_refresh_hash
ON auth_sessions (refresh_token_hash);

-- ตารางเก็บการ reset password (otp หรือโค้ดชั่วคราว)
create table if not exists password_resets (
    reset_pass_id         serial primary key,                      -- id auto increment
    reset_pass_user_id    integer references users(user_id) 
                          on delete cascade,                       -- ผูกกับ users
    otp_hash              varchar(255) not null,                   -- เก็บรหัส OTP แบบ hash
    reset_pass_expires_at timestamptz not null,                    -- เวลาหมดอายุของการ reset
    used_at               timestamptz                              -- เวลาใช้ reset ไปแล้ว
);
-- ตารางยืนยันอีเมลก่อนสมัครสมาชิก (OTP) 888
create table if not exists email_verifications (
    verify_id         serial primary key,
    email             varchar(256) not null,
    otp_hash          varchar(255) not null,
    expires_at        timestamptz not null,
    used_at           timestamptz,
    created_at        timestamptz default now()
);

-- กัน query ช้า + กันเคสตัวเล็กใหญ่
create index if not exists ix_email_verifications_email_ci
  on email_verifications (lower(email));

create index if not exists ix_email_verifications_active
  on email_verifications (lower(email), expires_at)
  where used_at is null;

--------------------------------------------------------------------------------------------------------------
-- เพิ่มตารางหัวข้อให้มาก่อน user_interests
create table if not exists topics (
    topic_id   serial primary key,
    topic_name varchar(20) not null unique
);

-- ตารางเก็บความสนใจของผู้ใช้ (mapping user ↔ topic)
create table if not exists user_interests (
    interest_user_id    integer references users(user_id) on delete cascade,   -- ผู้ใช้
    interest_topic_id   integer references topics(topic_id) on delete cascade, -- หัวข้อที่สนใจ
    interest_created_at timestamptz default now(),                             -- เวลาเพิ่มความสนใจ
    primary key (interest_user_id, interest_topic_id)                          -- กันซ้ำ user เลือก topic เดิม
);
----------------------------------------------------------------------------------------------------------------
-- ตารางเก็บการติดตาม (Follow)
create table if not exists follows (
    follower_user_id   integer references users(user_id) on delete cascade, -- คนที่กดติดตาม
    followed_user_id   integer references users(user_id) on delete cascade, -- คนที่ถูกติดตาม
    follow_created_at  timestamptz default now(),                           -- เวลา follow
    primary key (follower_user_id, followed_user_id),                        -- กันซ้ำ
    check (follower_user_id <> followed_user_id)
);
create index if not exists ix_follows_follower on follows(follower_user_id);
create index if not exists ix_follows_followed on follows(followed_user_id);

-- enum สถานะคำขอเป็นเพื่อน (ใช้ DO-block กันกรณี IF NOT EXISTS ใช้ไม่ได้)
do $$
begin
  if not exists (select 1 from pg_type where typname = 'friend_request_status') then
    create type friend_request_status as enum ('pending','accepted','declined');
  end if;
end
$$ language plpgsql;

-- ตารางเก็บคำขอเป็นเพื่อน
create table if not exists friend_requests (
    request_id         serial primary key,                                      -- id auto increment
    requester_user_id  integer not null references users(user_id) on delete cascade, -- คนที่ส่งคำขอ
    addressee_user_id  integer not null references users(user_id) on delete cascade, -- คนที่ถูกส่งคำขอ
    request_status     friend_request_status not null default 'pending',        -- สถานะ
    request_created_at timestamptz default now(),                               -- เวลาเริ่มส่งคำขอ
    decided_at         timestamptz,                                             -- เวลาตอบรับ/ปฏิเสธ
    check (requester_user_id <> addressee_user_id)                              -- กันไม่ให้ส่งหาตัวเอง
);

-- index กัน pending ซ้ำทิศทาง (A → B, B → A)
create unique index if not exists uq_friend_requests_pending_pair
on friend_requests (
    least(requester_user_id, addressee_user_id),
    greatest(requester_user_id, addressee_user_id)
) where request_status = 'pending';

create index if not exists ix_friendreq_addressee_pending
  on friend_requests(addressee_user_id) where request_status='pending';
create index if not exists ix_friendreq_requester
  on friend_requests(requester_user_id);

-- ตารางเพื่อน (friendships) ที่ถูก accept แล้ว
create table if not exists friendships (
    user_id    integer not null references users(user_id) on delete cascade, -- user
    friend_id  integer not null references users(user_id) on delete cascade, -- friend
    created_at timestamptz default now(),                                    -- เวลาเป็นเพื่อน
    primary key (user_id, friend_id),
    check (user_id <> friend_id),                                            -- กันไม่ให้เป็นเพื่อนกับตัวเอง
    check (user_id < friend_id)                                              -- เก็บทิศเดียว (user_id < friend_id)
);
create index if not exists ix_friendships_user   on friendships(user_id);
create index if not exists ix_friendships_friend on friendships(friend_id);

-- ตารางเก็บไฟล์เอกสาร (documents)
create table if not exists documents (
    document_id       serial primary key,
    document_user_id  integer references users(user_id) on delete cascade, -- เจ้าของไฟล์
    document_name     varchar(255),                                        -- ชื่อไฟล์
    document_url      text not null,                                       -- URL ของไฟล์
    storage_provider  varchar(50),                                         -- เช่น s3, firebase, local
    uploaded_at       timestamptz default now()                            -- เวลาอัปโหลด
);

-- ตารางเก็บสรุป (summaries)
create table if not exists summaries (
    summary_id            serial primary key,
    summary_document_id   integer not null references documents(document_id) on delete cascade,
    summary_status        varchar(20) not null default 'queued' check (summary_status in ('queued','processing','done','failed')),
    summary_task_id       text,
    summary_error_message text,
    summary_text          text,  -- ตัวสรุปข้อความ
    summary_html          text,           -- ถ้ามี highlight HTML
    summary_pdf_url       text,           -- ถ้ามี export PDF
    summary_created_at    timestamptz default now(), -- เวลา generate
    summary_started_at    timestamptz,
    summary_finished_at   timestamptz,
    summary_updated_at    timestamptz default now()
);

create index if not exists ix_summaries_document_id on summaries(summary_document_id);

-- ตารางโพสต์
create table if not exists posts (
    post_id             serial primary key,
    post_author_user_id integer references users(user_id) on delete cascade,  -- ผู้โพสต์
    post_title          varchar(120) not null,                                -- หัวข้อโพสต์
    post_description    text,                                                 -- คำอธิบาย
    post_visibility     varchar(10) not null check (post_visibility in ('public','friends')),    -- การมองเห็น
    post_document_id    integer references documents(document_id) on delete set null,            -- อ้างไฟล์
    post_cover_url      text,       
    post_created_at     timestamptz default now(),
    post_updated_at     timestamptz default now()
);

create index if not exists ix_posts_document_id on posts(post_document_id);

-- ตารางแท็ก (tags)
create table if not exists tags (
    tag_id   serial primary key,
    tag_name varchar(50) unique not null -- ชื่อแท็กไม่ซ้ำ
);

-- ตารางเชื่อมโพสต์กับแท็ก (many-to-many)
create table if not exists post_tags (
    post_tag_post_id integer references posts(post_id) on delete cascade,
    post_tag_tag_id  integer references tags(tag_id) on delete cascade,
    primary key (post_tag_post_id, post_tag_tag_id)
);

-- ตารางเก็บการกดถูกใจ
create table if not exists likes (
    like_user_id    integer references users(user_id) on delete cascade, -- คนที่กดไลก์
    like_post_id    integer references posts(post_id) on delete cascade, -- โพสต์
    like_created_at timestamptz default now(),                           -- เวลาไลก์
    primary key (like_user_id, like_post_id)
);

-- ตารางเก็บการบันทึกโพสต์
create table if not exists saved_posts (
    save_user_id    integer references users(user_id) on delete cascade, -- คนที่บันทึก
    save_post_id    integer references posts(post_id) on delete cascade, -- โพสต์
    save_created_at timestamptz default now(),                           -- เวลาบันทึก
    primary key (save_user_id, save_post_id)
);

-- ตารางสถิติโพสต์
create table if not exists post_stats (
    post_stats_post_id    integer primary key references posts(post_id) on delete cascade, -- อิงโพสต์
    post_like_count       integer default 0,    -- จำนวนไลก์
    post_save_count       integer default 0,    -- จำนวนบันทึก
    post_last_activity_at timestamptz default now() -- เวลากิจกรรมล่าสุด
);


CREATE INDEX IF NOT EXISTS idx_likes_post_id
ON likes (like_post_id);

CREATE INDEX IF NOT EXISTS idx_saved_posts_post_id
ON saved_posts (save_post_id);


CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger AS $$
BEGIN
  NEW.updated_at := now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ตาราง document_features
CREATE TABLE IF NOT EXISTS document_features (
  document_id      integer PRIMARY KEY
                  REFERENCES documents(document_id) ON DELETE CASCADE,
  feature_status   varchar(20) NOT NULL DEFAULT 'queued'
                  CHECK (feature_status IN ('queued','processing','done','failed')),
  style_label      varchar(20)
                  CHECK (style_label IN ('typed','handwritten','empty','not_typed','unknown')),
  style_vector_v16 vector(16),
  style_vector_raw jsonb,
  cluster_id       integer CHECK (cluster_id IS NULL OR cluster_id >= -1),
  cluster_updated_at timestamptz,
  content_text     text,
  content_embedding vector(768),
  classify_debug   jsonb, -- debug จาก classify
  process_trace    jsonb, -- trace ละเอียดจาก process_one_pdf 
  error_message    text,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now()
);

-- Trigger อัปเดต updated_at
DROP TRIGGER IF EXISTS trg_document_features_updated_at ON document_features;
CREATE TRIGGER trg_document_features_updated_at
BEFORE UPDATE ON document_features
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS ix_document_features_status
  ON document_features(feature_status);

CREATE INDEX IF NOT EXISTS ix_document_features_style_label
  ON document_features(style_label);

CREATE INDEX IF NOT EXISTS ix_document_features_cluster_id
  ON document_features(cluster_id);

-- (แนะนำเพิ่ม) เร็วขึ้นเวลาคัด typed/hand + cluster เดียวกัน
CREATE INDEX IF NOT EXISTS ix_document_features_label_cluster
  ON document_features(style_label, cluster_id);

-- pgvector HNSW index (style vector)
CREATE INDEX IF NOT EXISTS ix_document_features_stylevec_hnsw
  ON document_features
  USING hnsw (style_vector_v16 vector_cosine_ops)
  WHERE style_vector_v16 IS NOT NULL;

CREATE TABLE IF NOT EXISTS recommendations (
  rec_user_id   integer references users(user_id) on delete cascade,
  rec_post_id   integer references posts(post_id) on delete cascade,
  score         real not null,
  seed_document_id integer,
  created_at    timestamptz default now(),
  primary key (rec_user_id, rec_post_id)
);

CREATE INDEX IF NOT EXISTS ix_recommendations_user_time
ON recommendations (rec_user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS ix_recommendations_user_score_time
ON recommendations (rec_user_id, score DESC, created_at DESC);

commit;