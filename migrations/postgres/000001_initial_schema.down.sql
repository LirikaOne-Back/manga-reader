DROP TABLE IF EXISTS pages;
DROP TABLE IF EXISTS chapters;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS manga;

DROP INDEX IF EXISTS idx_manga_title;
DROP INDEX IF EXISTS idx_chapters_manga_id;
DROP INDEX IF EXISTS idx_pages_chapter_id;
DROP INDEX IF EXISTS idx_users_username;