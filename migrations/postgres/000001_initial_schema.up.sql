CREATE TABLE IF NOT EXISTS manga (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    DESCRIPTION TEXT
);

CREATE TABLE IF NOT EXISTS chapters (
    id SERIAL PRIMARY KEY,
    manga_id INTEGER NOT NULL,
    number INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    CONSTRAINT fk_manga FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS pages (
    id SERIAL PRIMARY KEY,
    chapter_id INTEGER NOT NULL,
    number INTEGER NOT NULL,
    image_path VARCHAR(255) NOT NULL,
    CONSTRAINT fk_chapter FOREIGN KEY (chapter_id) REFERENCES chapters(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL
);

CREATE INDEX idx_manga_title ON manga(title);
CREATE INDEX idx_chapters_manga_id ON chapters(manga_id);
CREATE INDEX idx_pages_chapter_id ON pages(chapter_id);
CREATE INDEX idx_users_username ON users(username);