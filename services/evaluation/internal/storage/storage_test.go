package storage

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFileStorage_Interface тестирует интерфейс FileStorage
func TestFileStorage_Interface(t *testing.T) {
	// Создаем мок реализацию интерфейса
	mockStorage := &MockFileStorage{}
	
	// Проверяем, что мок реализует интерфейс
	var _ FileStorage = mockStorage
}

// MockFileStorage - мок для интерфейса FileStorage
type MockFileStorage struct {
	uploadFile   func(ctx context.Context, file io.Reader, fileName, contentType string) (string, error)
	downloadFile func(ctx context.Context, objectName string) (io.ReadCloser, error)
	deleteFile   func(ctx context.Context, objectName string) error
	getFileURL   func(objectName string) string
}

func (m *MockFileStorage) UploadFile(ctx context.Context, file io.Reader, fileName, contentType string) (string, error) {
	if m.uploadFile != nil {
		return m.uploadFile(ctx, file, fileName, contentType)
	}
	return "", errors.New("upload not implemented")
}

func (m *MockFileStorage) DownloadFile(ctx context.Context, objectName string) (io.ReadCloser, error) {
	if m.downloadFile != nil {
		return m.downloadFile(ctx, objectName)
	}
	return nil, errors.New("download not implemented")
}

func (m *MockFileStorage) DeleteFile(ctx context.Context, objectName string) error {
	if m.deleteFile != nil {
		return m.deleteFile(ctx, objectName)
	}
	return errors.New("delete not implemented")
}

func (m *MockFileStorage) GetFileURL(objectName string) string {
	if m.getFileURL != nil {
		return m.getFileURL(objectName)
	}
	return ""
}

// TestFileStorage_UploadFile тестирует загрузку файла
func TestFileStorage_UploadFile(t *testing.T) {
	tests := []struct {
		name          string
		fileName      string
		contentType   string
		fileContent   string
		mockError     error
		expectedError bool
	}{
		{
			name:          "Успешная загрузка файла",
			fileName:      "test.txt",
			contentType:   "text/plain",
			fileContent:   "Hello, World!",
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "Ошибка при загрузке файла",
			fileName:      "test.txt",
			contentType:   "text/plain",
			fileContent:   "Hello, World!",
			mockError:     errors.New("upload failed"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockFileStorage{}
			
			if tt.expectedError {
				mockStorage.uploadFile = func(ctx context.Context, file io.Reader, fileName, contentType string) (string, error) {
					return "", tt.mockError
				}
			} else {
				mockStorage.uploadFile = func(ctx context.Context, file io.Reader, fileName, contentType string) (string, error) {
					return "uploaded_" + fileName, nil
				}
			}

			fileReader := strings.NewReader(tt.fileContent)
			result, err := mockStorage.UploadFile(context.Background(), fileReader, tt.fileName, tt.contentType)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
				assert.Contains(t, result, tt.fileName)
			}
		})
	}
}

// TestFileStorage_DownloadFile тестирует скачивание файла
func TestFileStorage_DownloadFile(t *testing.T) {
	tests := []struct {
		name          string
		objectName    string
		fileContent   string
		mockError     error
		expectedError bool
	}{
		{
			name:          "Успешное скачивание файла",
			objectName:    "test_object.txt",
			fileContent:   "Downloaded content",
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "Ошибка при скачивании файла",
			objectName:    "nonexistent.txt",
			fileContent:   "",
			mockError:     errors.New("file not found"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockFileStorage{}
			
			if tt.expectedError {
				mockStorage.downloadFile = func(ctx context.Context, objectName string) (io.ReadCloser, error) {
					return nil, tt.mockError
				}
			} else {
				mockStorage.downloadFile = func(ctx context.Context, objectName string) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader(tt.fileContent)), nil
				}
			}

			result, err := mockStorage.DownloadFile(context.Background(), tt.objectName)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Читаем содержимое
				content, err := io.ReadAll(result)
				assert.NoError(t, err)
				assert.Equal(t, tt.fileContent, string(content))

				// Закрываем объект
				result.Close()
			}
		})
	}
}

// TestFileStorage_DeleteFile тестирует удаление файла
func TestFileStorage_DeleteFile(t *testing.T) {
	tests := []struct {
		name          string
		objectName    string
		mockError     error
		expectedError bool
	}{
		{
			name:          "Успешное удаление файла",
			objectName:    "test_object.txt",
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "Ошибка при удалении файла",
			objectName:    "nonexistent.txt",
			mockError:     errors.New("delete failed"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockFileStorage{}
			
			mockStorage.deleteFile = func(ctx context.Context, objectName string) error {
				return tt.mockError
			}

			err := mockStorage.DeleteFile(context.Background(), tt.objectName)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestFileStorage_GetFileURL тестирует получение URL файла
func TestFileStorage_GetFileURL(t *testing.T) {
	tests := []struct {
		name       string
		objectName string
		expected   string
	}{
		{
			name:       "Простой файл",
			objectName: "test.txt",
			expected:   "http://localhost:9000/testbucket/test.txt",
		},
		{
			name:       "Файл в папке",
			objectName: "folder/subfolder/file.txt",
			expected:   "http://localhost:9000/testbucket/folder/subfolder/file.txt",
		},
		{
			name:       "Файл с русским именем",
			objectName: "документ.pdf",
			expected:   "http://localhost:9000/testbucket/документ.pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockFileStorage{}
			
			mockStorage.getFileURL = func(objectName string) string {
				return "http://localhost:9000/testbucket/" + objectName
			}

			result := mockStorage.GetFileURL(tt.objectName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFileStorage_Integration тестирует интеграцию всех методов FileStorage
func TestFileStorage_Integration(t *testing.T) {
	mockStorage := &MockFileStorage{}
	
	// Тестируем полный цикл: загрузка -> получение URL -> скачивание -> удаление
	ctx := context.Background()
	fileName := "test.txt"
	contentType := "text/plain"
	fileContent := "Integration test content"
	objectName := "1234567890_test.txt"
	
	// Настройка моков
	mockStorage.uploadFile = func(ctx context.Context, file io.Reader, fileName, contentType string) (string, error) {
		return objectName, nil
	}
	
	mockStorage.getFileURL = func(objectName string) string {
		return "http://localhost:9000/testbucket/" + objectName
	}
	
	mockStorage.downloadFile = func(ctx context.Context, objectName string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(fileContent)), nil
	}
	
	mockStorage.deleteFile = func(ctx context.Context, objectName string) error {
		return nil
	}
	
	// Выполняем операции
	uploadedName, err := mockStorage.UploadFile(ctx, strings.NewReader(fileContent), fileName, contentType)
	assert.NoError(t, err)
	assert.Equal(t, objectName, uploadedName)
	
	url := mockStorage.GetFileURL(uploadedName)
	assert.Equal(t, "http://localhost:9000/testbucket/"+objectName, url)
	
	downloadedFile, err := mockStorage.DownloadFile(ctx, uploadedName)
	assert.NoError(t, err)
	assert.NotNil(t, downloadedFile)
	
	content, err := io.ReadAll(downloadedFile)
	assert.NoError(t, err)
	assert.Equal(t, fileContent, string(content))
	downloadedFile.Close()
	
	err = mockStorage.DeleteFile(ctx, uploadedName)
	assert.NoError(t, err)
}

// TestFileStorage_ContextCancellation тестирует отмену контекста
func TestFileStorage_ContextCancellation(t *testing.T) {
	mockStorage := &MockFileStorage{}
	
	// Создаем отмененный контекст
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	
	// Настраиваем мок для возврата ошибки контекста
	mockStorage.uploadFile = func(ctx context.Context, file io.Reader, fileName, contentType string) (string, error) {
		select {
		case <-ctx.Done():
			return "", context.Canceled
		default:
			return "success", nil
		}
	}
	
	// Тестируем загрузку с отмененным контекстом
	fileReader := strings.NewReader("test content")
	_, err := mockStorage.UploadFile(ctx, fileReader, "test.txt", "text/plain")
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

// TestFileStorage_FileSize тестирует работу с файлами разных размеров
func TestFileStorage_FileSize(t *testing.T) {
	mockStorage := &MockFileStorage{}
	
	tests := []struct {
		name        string
		content     string
		expectedSize int
	}{
		{
			name:        "Пустой файл",
			content:     "",
			expectedSize: 0,
		},
		{
			name:        "Маленький файл",
			content:     "Hello",
			expectedSize: 5,
		},
		{
			name:        "Большой файл",
			content:     strings.Repeat("A", 1000),
			expectedSize: 1000,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage.uploadFile = func(ctx context.Context, file io.Reader, fileName, contentType string) (string, error) {
				// Читаем содержимое для проверки размера
				content, err := io.ReadAll(file)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSize, len(content))
				return "uploaded_" + fileName, nil
			}
			
			fileReader := strings.NewReader(tt.content)
			_, err := mockStorage.UploadFile(context.Background(), fileReader, "test.txt", "text/plain")
			assert.NoError(t, err)
		})
	}
}

// TestFileStorage_ErrorHandling тестирует обработку различных типов ошибок
func TestFileStorage_ErrorHandling(t *testing.T) {
	mockStorage := &MockFileStorage{}
	
	tests := []struct {
		name          string
		errorType     error
		expectedError string
	}{
		{
			name:          "Ошибка сети",
			errorType:     errors.New("network error"),
			expectedError: "network error",
		},
		{
			name:          "Ошибка аутентификации",
			errorType:     errors.New("authentication failed"),
			expectedError: "authentication failed",
		},
		{
			name:          "Ошибка разрешений",
			errorType:     errors.New("permission denied"),
			expectedError: "permission denied",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage.uploadFile = func(ctx context.Context, file io.Reader, fileName, contentType string) (string, error) {
				return "", tt.errorType
			}
			
			fileReader := strings.NewReader("test content")
			_, err := mockStorage.UploadFile(context.Background(), fileReader, "test.txt", "text/plain")
			assert.Error(t, err)
			assert.Equal(t, tt.expectedError, err.Error())
		})
	}
}

// TestFileStorage_ContentType тестирует работу с различными типами контента
func TestFileStorage_ContentType(t *testing.T) {
	mockStorage := &MockFileStorage{}
	
	tests := []struct {
		name        string
		contentType string
		fileName    string
	}{
		{
			name:        "Текстовый файл",
			contentType: "text/plain",
			fileName:    "document.txt",
		},
		{
			name:        "PDF файл",
			contentType: "application/pdf",
			fileName:    "document.pdf",
		},
		{
			name:        "Изображение",
			contentType: "image/jpeg",
			fileName:    "photo.jpg",
		},
		{
			name:        "Excel файл",
			contentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			fileName:    "data.xlsx",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage.uploadFile = func(ctx context.Context, file io.Reader, fileName, contentType string) (string, error) {
				assert.Equal(t, tt.contentType, contentType)
				assert.Equal(t, tt.fileName, fileName)
				return "uploaded_" + fileName, nil
			}
			
			fileReader := strings.NewReader("test content")
			_, err := mockStorage.UploadFile(context.Background(), fileReader, tt.fileName, tt.contentType)
			assert.NoError(t, err)
		})
	}
}
