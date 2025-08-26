BEGIN;

-- Удаляем таблицы в обратном порядке (из-за внешних ключей)
DROP TABLE IF EXISTS remarks;
DROP TABLE IF EXISTS project_files;
DROP TABLE IF EXISTS projects;

-- Удаляем enum типы
DROP TYPE IF EXISTS file_type;
DROP TYPE IF EXISTS project_status;

COMMIT;
