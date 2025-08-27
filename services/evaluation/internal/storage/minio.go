package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOStorage реализация FileStorage для MinIO
type MinIOStorage struct {
	client     *minio.Client
	bucketName string
	endpoint   string
}

// NewMinIOStorage создает новый экземпляр MinIOStorage
func NewMinIOStorage(endpoint, accessKey, secretKey, bucketName string, useSSL bool) (*MinIOStorage, error) {
	// Создаем клиент MinIO
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Проверяем существование bucket
	exists, err := minioClient.BucketExists(context.Background(), bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	// Создаем bucket если не существует
	if !exists {
		err = minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &MinIOStorage{
		client:     minioClient,
		bucketName: bucketName,
		endpoint:   endpoint,
	}, nil
}

// UploadFile загружает файл в MinIO
func (m *MinIOStorage) UploadFile(ctx context.Context, file io.Reader, fileName, contentType string) (string, error) {
	// Генерируем уникальное имя объекта
	objectName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(fileName))

	// Загружаем файл
	_, err := m.client.PutObject(ctx, m.bucketName, objectName, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return objectName, nil
}

// DownloadFile скачивает файл из MinIO
func (m *MinIOStorage) DownloadFile(ctx context.Context, objectName string) (io.ReadCloser, error) {
	// Добавляем логирование для отладки
	fmt.Printf("Attempting to download file: bucket=%s, object=%s\n", m.bucketName, objectName)

	obj, err := m.client.GetObject(ctx, m.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return obj, nil
}

// DeleteFile удаляет файл из MinIO
func (m *MinIOStorage) DeleteFile(ctx context.Context, objectName string) error {
	err := m.client.RemoveObject(ctx, m.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetFileURL возвращает URL для доступа к файлу
func (m *MinIOStorage) GetFileURL(objectName string) string {
	protocol := "http"
	if m.client.EndpointURL().Scheme == "https" {
		protocol = "https"
	}

	return fmt.Sprintf("%s://%s/%s/%s", protocol, m.endpoint, m.bucketName, objectName)
}
