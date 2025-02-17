CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS profiles (
    user_id INTEGER PRIMARY KEY REFERENCES users(id),
    name TEXT,
    bio TEXT,
    profile_picture TEXT,
    location TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_bios (
    user_id INTEGER PRIMARY KEY REFERENCES users(id),
    interests TEXT[],
    hobbies TEXT[],
    music_preferences TEXT[],
    food_preferences TEXT[],
    looking_for TEXT[],
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS connections (
    id SERIAL PRIMARY KEY,
    user_id_1 INTEGER REFERENCES users(id),
    user_id_2 INTEGER REFERENCES users(id),
    status TEXT DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id_1, user_id_2)
);

CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    connection_id INTEGER REFERENCES connections(id),
    sender_id INTEGER REFERENCES users(id),
    content TEXT NOT NULL,
    read BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW()
);