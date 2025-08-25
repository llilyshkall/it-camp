package storage

import (
	"context"
	"io"
)

// FileStorage интерфейс для работы с файловым хранилищем
type FileStorage interface {
	// UploadFile загружает файл в хранилище
	UploadFile(ctx context.Context, file io.Reader, fileName, contentType string) (string, error)

	// DownloadFile скачивает файл из хранилища
	DownloadFile(ctx context.Context, objectName string) (io.ReadCloser, error)

	// DeleteFile удаляет файл из хранилища
	DeleteFile(ctx context.Context, objectName string) error

	// GetFileURL возвращает URL для доступа к файлу
	GetFileURL(objectName string) string
}
