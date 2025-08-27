package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMinIOStorage_Integration тестирует интеграцию с реальным MinIO сервером
// Эти тесты требуют запущенного MinIO сервера
func TestMinIOStorage_Integration(t *testing.T) {
	// Пропускаем тесты, если нет MinIO сервера
	t.Skip("Требует запущенного MinIO сервера")

	// Конфигурация для тестового MinIO сервера
	endpoint := "localhost:9000"
	accessKey := "minioadmin"
	secretKey := "minioadmin"
	bucketName := "test-bucket"
	useSSL := false

	// Создаем хранилище
	storage, err := NewMinIOStorage(endpoint, accessKey, secretKey, bucketName, useSSL)
	require.NoError(t, err)
	require.NotNil(t, storage)

	t.Run("Полный цикл работы с файлами", func(t *testing.T) {
		ctx := context.Background()
		fileName := "test_integration.txt"
		contentType := "text/plain"
		fileContent := "Integration test content " + time.Now().String()

		// 1. Загружаем файл
		fileReader := strings.NewReader(fileContent)
		objectName, err := storage.UploadFile(ctx, fileReader, fileName, contentType)
		require.NoError(t, err)
		require.NotEmpty(t, objectName)
		assert.Contains(t, objectName, fileName)

		// 2. Получаем URL файла
		fileURL := storage.GetFileURL(objectName)
		require.NotEmpty(t, fileURL)
		assert.Contains(t, fileURL, bucketName)
		assert.Contains(t, fileURL, objectName)

		// 3. Скачиваем файл
		downloadedFile, err := storage.DownloadFile(ctx, objectName)
		require.NoError(t, err)
		require.NotNil(t, downloadedFile)

		// Читаем содержимое
		content, err := io.ReadAll(downloadedFile)
		require.NoError(t, err)
		assert.Equal(t, fileContent, string(content))

		// Закрываем файл
		downloadedFile.Close()

		// 4. Удаляем файл
		err = storage.DeleteFile(ctx, objectName)
		require.NoError(t, err)

		// 5. Проверяем, что файл удален
		_, err = storage.DownloadFile(ctx, objectName)
		assert.Error(t, err)
	})

	t.Run("Загрузка файлов разных типов", func(t *testing.T) {
		ctx := context.Background()

		tests := []struct {
			name        string
			content     string
			contentType string
			fileName    string
		}{
			{
				name:        "Текстовый файл",
				content:     "Hello, World!",
				contentType: "text/plain",
				fileName:    "text.txt",
			},
			{
				name:        "JSON файл",
				content:     `{"key": "value"}`,
				contentType: "application/json",
				fileName:    "data.json",
			},
			{
				name:        "HTML файл",
				content:     "<html><body>Test</body></html>",
				contentType: "text/html",
				fileName:    "page.html",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				fileReader := strings.NewReader(tt.content)
				objectName, err := storage.UploadFile(ctx, fileReader, tt.fileName, tt.contentType)
				require.NoError(t, err)
				require.NotEmpty(t, objectName)

				// Проверяем, что файл загружен
				downloadedFile, err := storage.DownloadFile(ctx, objectName)
				require.NoError(t, err)

				content, err := io.ReadAll(downloadedFile)
				require.NoError(t, err)
				assert.Equal(t, tt.content, string(content))
				downloadedFile.Close()

				// Удаляем тестовый файл
				err = storage.DeleteFile(ctx, objectName)
				require.NoError(t, err)
			})
		}
	})

	t.Run("Работа с большими файлами", func(t *testing.T) {
		ctx := context.Background()
		fileName := "large_file.txt"
		contentType := "text/plain"

		// Создаем большой файл (1MB)
		largeContent := strings.Repeat("A", 1024*1024)
		fileReader := strings.NewReader(largeContent)

		objectName, err := storage.UploadFile(ctx, fileReader, fileName, contentType)
		require.NoError(t, err)
		require.NotEmpty(t, objectName)

		// Скачиваем и проверяем
		downloadedFile, err := storage.DownloadFile(ctx, objectName)
		require.NoError(t, err)

		content, err := io.ReadAll(downloadedFile)
		require.NoError(t, err)
		assert.Equal(t, len(largeContent), len(content))
		assert.Equal(t, largeContent, string(content))
		downloadedFile.Close()

		// Удаляем файл
		err = storage.DeleteFile(ctx, objectName)
		require.NoError(t, err)
	})

	t.Run("Обработка ошибок", func(t *testing.T) {
		ctx := context.Background()

		// Попытка скачать несуществующий файл
		_, err := storage.DownloadFile(ctx, "nonexistent_file.txt")
		assert.Error(t, err)

		// Попытка удалить несуществующий файл
		err = storage.DeleteFile(ctx, "nonexistent_file.txt")
		assert.Error(t, err)
	})
}

// TestMinIOStorage_Concurrent тестирует конкурентную работу с хранилищем
func TestMinIOStorage_Concurrent(t *testing.T) {
	t.Skip("Требует запущенного MinIO сервера")

	endpoint := "localhost:9000"
	accessKey := "minioadmin"
	secretKey := "minioadmin"
	bucketName := "test-bucket"
	useSSL := false

	storage, err := NewMinIOStorage(endpoint, accessKey, secretKey, bucketName, useSSL)
	require.NoError(t, err)

	t.Run("Конкурентная загрузка файлов", func(t *testing.T) {
		ctx := context.Background()
		numFiles := 10
		results := make(chan string, numFiles)
		errors := make(chan error, numFiles)

		// Запускаем горутины для загрузки файлов
		for i := 0; i < numFiles; i++ {
			go func(index int) {
				fileName := fmt.Sprintf("concurrent_file_%d.txt", index)
				content := fmt.Sprintf("Content for file %d", index)
				fileReader := strings.NewReader(content)

				objectName, err := storage.UploadFile(ctx, fileReader, fileName, "text/plain")
				if err != nil {
					errors <- err
					return
				}
				results <- objectName
			}(i)
		}

		// Собираем результаты
		uploadedFiles := make([]string, 0, numFiles)
		for i := 0; i < numFiles; i++ {
			select {
			case objectName := <-results:
				uploadedFiles = append(uploadedFiles, objectName)
			case err := <-errors:
				t.Errorf("Ошибка при загрузке файла: %v", err)
			case <-time.After(30 * time.Second):
				t.Fatal("Таймаут при конкурентной загрузке")
			}
		}

		// Проверяем, что все файлы загружены
		assert.Len(t, uploadedFiles, numFiles)

		// Удаляем все загруженные файлы
		for _, objectName := range uploadedFiles {
			err := storage.DeleteFile(ctx, objectName)
			require.NoError(t, err)
		}
	})
}

// TestMinIOStorage_Performance тестирует производительность хранилища
func TestMinIOStorage_Performance(t *testing.T) {
	t.Skip("Требует запущенного MinIO сервера")

	endpoint := "localhost:9000"
	accessKey := "minioadmin"
	secretKey := "minioadmin"
	bucketName := "test-bucket"
	useSSL := false

	storage, err := NewMinIOStorage(endpoint, accessKey, secretKey, bucketName, useSSL)
	require.NoError(t, err)

	t.Run("Измерение времени загрузки", func(t *testing.T) {
		ctx := context.Background()
		fileName := "performance_test.txt"
		content := strings.Repeat("Performance test content ", 1000) // ~25KB

		start := time.Now()
		fileReader := strings.NewReader(content)
		objectName, err := storage.UploadFile(ctx, fileReader, fileName, "text/plain")
		uploadTime := time.Since(start)

		require.NoError(t, err)
		t.Logf("Время загрузки файла ~25KB: %v", uploadTime)

		// Очистка
		err = storage.DeleteFile(ctx, objectName)
		require.NoError(t, err)
	})

	t.Run("Измерение времени скачивания", func(t *testing.T) {
		ctx := context.Background()
		fileName := "download_performance_test.txt"
		content := strings.Repeat("Download performance test content ", 1000)

		// Загружаем файл
		fileReader := strings.NewReader(content)
		objectName, err := storage.UploadFile(ctx, fileReader, fileName, "text/plain")
		require.NoError(t, err)

		// Измеряем время скачивания
		start := time.Now()
		downloadedFile, err := storage.DownloadFile(ctx, objectName)
		downloadTime := time.Since(start)

		require.NoError(t, err)
		t.Logf("Время скачивания файла ~25KB: %v", downloadTime)

		// Проверяем содержимое
		downloadedContent, err := io.ReadAll(downloadedFile)
		require.NoError(t, err)
		assert.Equal(t, content, string(downloadedContent))
		downloadedFile.Close()

		// Очистка
		err = storage.DeleteFile(ctx, objectName)
		require.NoError(t, err)
	})
}

// TestMinIOStorage_ErrorScenarios тестирует различные сценарии ошибок
func TestMinIOStorage_ErrorScenarios(t *testing.T) {
	t.Skip("Требует запущенного MinIO сервера")

	endpoint := "localhost:9000"
	accessKey := "minioadmin"
	secretKey := "minioadmin"
	bucketName := "test-bucket"
	useSSL := false

	_, err := NewMinIOStorage(endpoint, accessKey, secretKey, bucketName, useSSL)
	require.NoError(t, err)

	t.Run("Неправильные учетные данные", func(t *testing.T) {
		// Создаем хранилище с неправильными учетными данными
		_, err := NewMinIOStorage(endpoint, "wrong_key", "wrong_secret", bucketName, useSSL)
		assert.Error(t, err)
	})

	t.Run("Несуществующий endpoint", func(t *testing.T) {
		// Создаем хранилище с несуществующим endpoint
		_, err := NewMinIOStorage("nonexistent:9000", accessKey, secretKey, bucketName, useSSL)
		assert.Error(t, err)
	})

	t.Run("Пустое имя bucket", func(t *testing.T) {
		// Создаем хранилище с пустым именем bucket
		_, err := NewMinIOStorage(endpoint, accessKey, secretKey, "", useSSL)
		assert.Error(t, err)
	})
}

// TestMinIOStorage_SSL тестирует работу с SSL
func TestMinIOStorage_SSL(t *testing.T) {
	t.Skip("Требует MinIO сервер с SSL")

	endpoint := "localhost:9000"
	accessKey := "minioadmin"
	secretKey := "minioadmin"
	bucketName := "test-bucket"
	useSSL := true

	storage, err := NewMinIOStorage(endpoint, accessKey, secretKey, bucketName, useSSL)
	require.NoError(t, err)

	t.Run("HTTPS URL", func(t *testing.T) {
		fileName := "ssl_test.txt"
		content := "SSL test content"
		fileReader := strings.NewReader(content)

		objectName, err := storage.UploadFile(context.Background(), fileReader, fileName, "text/plain")
		require.NoError(t, err)

		fileURL := storage.GetFileURL(objectName)
		assert.True(t, strings.HasPrefix(fileURL, "https://"), "URL должен использовать HTTPS")

		// Очистка
		err = storage.DeleteFile(context.Background(), objectName)
		require.NoError(t, err)
	})
}
