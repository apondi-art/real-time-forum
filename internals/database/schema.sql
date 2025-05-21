CREATE TABLE IF NOT EXISTS users(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nickname TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    age  INTEGER  NOT NULL,
    gender  TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    password TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);



CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
    FOREIGN KEY (category_id) REFERENCES categories(id)
);

-- comments table to match your database functions
CREATE TABLE IF NOT EXISTS comments (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 user_id INTEGER NOT NULL,
 post_id INTEGER NOT NULL,
 content TEXT NOT NULL,
 created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
 FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
 FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);


    -- Add to your schema.sql
CREATE TABLE IF NOT EXISTS user_status (
    user_id INTEGER PRIMARY KEY,
    online BOOLEAN NOT NULL DEFAULT FALSE,
    last_seen DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Insert the predefined categories
INSERT OR IGNORE INTO categories (name) VALUES
    ('Sports'),
    ('Lifestyle'),
    ('Education'),
    ('Finance'),
    ('Music'),
    ('Culture'),
    ('Technology'),
    ('Health'),
    ('Travel'),
    ('Food');


    CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);