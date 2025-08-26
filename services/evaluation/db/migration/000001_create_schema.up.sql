BEGIN;

-- Создаем enum для статуса проекта
CREATE TYPE project_status AS ENUM (
    'ready',                    -- готов, отдаем все поля
    'processing_remarks',       -- обрабатываются замечания
    'processing_checklist',     -- генерация чек-листа
    'generating_final_report'   -- генерируется итоговый отчет
);

-- Создаем enum для типа файла
CREATE TYPE file_type AS ENUM (
    'documentation',            -- документация (для чеклиста)
    'checklist',                -- чек-лист
    'remarks',                  -- замечания
    'remarks_clustered',        -- замечания, сгруппированные по тематике
    'final_report'              -- итоговый отчет
);

-- Создаем таблицу projects с финальной структурой
CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    status project_status DEFAULT 'ready' NOT NULL
);

-- Создаем таблицу project_files с финальной структурой
CREATE TABLE project_files (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    extension VARCHAR(10) NOT NULL,
    file_type file_type NOT NULL,
    uploaded_at TIMESTAMP DEFAULT NOW() NOT NULL
);

-- Создаем таблицу remarks
CREATE TABLE remarks (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    direction VARCHAR(255) NOT NULL,
    section VARCHAR(255) NOT NULL,
    subsection VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW() NOT NULL
);

COMMIT;
