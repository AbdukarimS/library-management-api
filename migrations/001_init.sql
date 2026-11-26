CREATE TABLE IF NOT EXISTS books (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    author VARCHAR(255) NOT NULL,
    year INT NOT NULL CHECK (year >= 1450),
    genre VARCHAR(100) NOT NULL,
    total_copies INT NOT NULL CHECK (total_copies >= 1),
    available_copies INT NOT NULL CHECK (available_copies >= 0 AND available_copies <= total_copies)
);

CREATE TABLE IF NOT EXISTS readers (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    age INT NOT NULL CHECK (age >= 6),
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS loans (
    id SERIAL PRIMARY KEY,
    book_id INT NOT NULL REFERENCES books(id) ON DELETE RESTRICT,
    reader_id INT NOT NULL REFERENCES readers(id) ON DELETE RESTRICT,
    borrowed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    due_date TIMESTAMP WITH TIME ZONE NOT NULL,
    returned_at TIMESTAMP WITH TIME ZONE NULL,
    is_returned BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'USER' CHECK (role IN ('USER', 'ADMIN')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Начальные тестовые книги и читатели (чтобы база не была пустой)
INSERT INTO books (title, author, year, genre, total_copies, available_copies)
VALUES
    ('Clean Code', 'Robert Martin', 2008, 'IT', 3, 3),
    ('The Go Programming Language', 'Alan Donovan', 2015, 'IT', 2, 2),
    ('Путь Абая', 'Мухтар Ауэзов', 1942, 'Классика', 4, 4);

INSERT INTO readers (first_name, last_name, email, age, is_active)
VALUES
    ('Алихан', 'Сериков', 'alikhan@example.com', 21, TRUE),
    ('Айгерим', 'Нурланова', 'aigerim@example.com', 20, TRUE)
ON CONFLICT (email) DO NOTHING;
