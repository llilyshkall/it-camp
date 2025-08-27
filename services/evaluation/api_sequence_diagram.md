# API Sequence Diagram - Evaluation Service

Диаграмма последовательностей для всех API endpoints сервиса evaluation.

## Диаграмма

```mermaid
sequenceDiagram
    participant Client as Клиент
    participant Service as Evaluation Service
    participant DB as База данных
    participant Storage as MinIO Storage
    participant ML as ML Service

    Note over Client,ML: Health Check
    Client->>Service: GET /health
    Service->>DB: CheckHealth()
    DB-->>Service: Health status
    Service-->>Client: 200 OK + health info

    Note over Client,ML: Project Management
    Client->>Service: GET /api/projects
    Service->>DB: ListProjects()
    DB-->>Service: Projects list
    Service-->>Client: 200 OK + projects

    Client->>Service: POST /api/projects
    Note right of Client: {"name": "Project Name"}
    Service->>DB: CreateProject(name)
    DB-->>Service: Created project
    Service-->>Client: 201 Created + project

    Client->>Service: GET /api/projects/{id}
    Service->>DB: GetProject(id)
    DB-->>Service: Project data
    Service-->>Client: 200 OK + project

    Note over Client,ML: Documentation Upload
    Client->>Service: POST /api/projects/{id}/documentation
    Note right of Client: multipart/form-data with file
    Service->>Storage: UploadDocumentation(projectID, file)
    Storage-->>Service: File uploaded
    Service->>DB: Save file metadata
    DB-->>Service: File record
    Service-->>Client: 202 Accepted + file info

    Note over Client,ML: Checklist Generation
    Client->>Service: POST /api/projects/{id}/checklist
    Service->>DB: Validate project status
    DB-->>Service: Project status
    Service->>Storage: GenerateChecklist(projectID)
    Storage-->>Service: Checklist generated
    Service-->>Client: 202 Accepted + message

    Client->>Service: GET /api/projects/{id}/checklist
    Service->>Storage: GetChecklist(projectID)
    Storage-->>Service: Checklist result
    Service-->>Client: 200 OK + checklist

    Note over Client,ML: Remarks Upload
    Client->>Service: POST /api/projects/{id}/remarks
    Note right of Client: multipart/form-data with file
    Service->>Storage: UploadRemarks(projectID, file)
    Storage-->>Service: File uploaded
    Service->>DB: Save file metadata
    DB-->>Service: File record
    Service-->>Client: 202 Accepted + file info

    Client->>Service: GET /api/projects/{id}/remarks_clustered
    Service->>Storage: GetRemarksClustered(projectID)
    Storage-->>Service: Clustered remarks
    Service-->>Client: 200 OK + clustered remarks

    Note over Client,ML: Final Report Generation
    Client->>Service: POST /api/projects/{id}/final_report
    Service->>DB: Validate project status
    DB-->>Service: Project status
    Service->>Storage: GenerateFinalReport(projectID)
    Storage-->>Service: Report generated
    Service-->>Client: 202 Accepted + message

    Client->>Service: GET /api/projects/{id}/final_report
    Service->>Storage: GetFinalReport(projectID)
    Storage-->>Service: Final report
    Service-->>Client: 200 OK + final report

    Note over Client,ML: Swagger Documentation
    Client->>Service: GET /api/docs/
    Service-->>Client: Swagger UI
```

## Описание API Endpoints

### 1. Health Check
- **GET** `/health` - Проверка состояния сервиса и подключения к БД

### 2. Project Management
- **GET** `/api/projects` - Получение списка всех проектов
- **POST** `/api/projects` - Создание нового проекта
- **GET** `/api/projects/{id}` - Получение проекта по ID

### 3. Documentation Upload
- **POST** `/api/projects/{id}/documentation` - Загрузка документации проекта (max 50MB)

### 4. Checklist Operations
- **POST** `/api/projects/{id}/checklist` - Запуск генерации чеклиста
- **GET** `/api/projects/{id}/checklist` - Получение результата чеклиста

### 5. Remarks Operations
- **POST** `/api/projects/{id}/remarks` - Загрузка файла замечаний (max 50MB)
- **GET** `/api/projects/{id}/remarks_clustered` - Получение кластеризованных замечаний

### 6. Final Report Operations
- **POST** `/api/projects/{id}/final_report` - Запуск генерации финального отчета
- **GET** `/api/projects/{id}/final_report` - Получение финального отчета

### 7. API Documentation
- **GET** `/api/docs/` - Swagger UI для API документации

## Компоненты системы

- **Client** - Клиентское приложение
- **Evaluation Service** - Основной сервис (объединяет все слои)
- **DB** - PostgreSQL база данных
- **Storage** - MinIO хранилище файлов
- **ML** - ML сервис для обработки данных

## HTTP Status Codes

- **200** - Успешный GET запрос
- **201** - Успешное создание ресурса
- **202** - Запрос принят в обработку
- **400** - Ошибка в запросе
- **404** - Ресурс не найден
- **409** - Конфликт (например, проект уже обрабатывается)
- **500** - Внутренняя ошибка сервера
