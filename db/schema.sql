CREATE TABLE IF NOT EXISTS "schema_migrations" (version varchar(128) primary key);
CREATE TABLE cache (
    k TEXT NOT NULL PRIMARY KEY,
    ttl INTEGER NOT NULL,
    data BLOB NOT NULL
) STRICT;
CREATE TABLE posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE CHECK(
        -- Cannot contain uppercase letters
        slug NOT LIKE '%[A-Z]%'
        -- Cannot contain special characters (except hyphen)
        AND slug NOT LIKE '%[^a-z0-9-]%'
        -- Cannot have consecutive hyphens
        AND slug NOT LIKE '%--+%'
        -- Cannot start or end with hyphen
        AND slug NOT LIKE '-%'
        AND slug NOT LIKE '%-'
        -- Cannot contain any whitespace (space, tab, newline)
        AND slug NOT LIKE '% %'
        AND slug NOT LIKE '%' || char(9) || '%'  -- Tab
        AND slug NOT LIKE '%' || char(10) || '%' -- Newline
        AND slug NOT LIKE '%' || char(13) || '%' -- Carriage return
        -- Length constraints
        AND length(slug) BETWEEN 3 AND 100
    ),
    content BLOB NOT NULL,
    meta_description TEXT NOT NULL,
    -- status TEXT CHECK(status IN ('draft', 'published')) NOT NULL DEFAULT 'draft',
    created_at TEXT NOT NULL
) STRICT;
CREATE TABLE artists (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
) STRICT;
CREATE TABLE albums (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    artist_id INTEGER,
    title TEXT NOT NULL,
    release_date TEXT NOT NULL,
    cover_url TEXT NOT NULL,
    FOREIGN KEY (artist_id) REFERENCES artists(id) ON DELETE CASCADE
) STRICT;
CREATE TABLE links (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    album_id INTEGER,
    platform TEXT NOT NULL,
    url TEXT NOT NULL,
    FOREIGN KEY (album_id) REFERENCES albums(id) ON DELETE CASCADE
) STRICT;
-- Dbmate schema migrations
INSERT INTO "schema_migrations" (version) VALUES
  ('20250131163110');
