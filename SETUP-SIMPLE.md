# Простая настройка деплоя без Kubernetes

## 🎯 **Что это дает:**

✅ **Автоматический деплой** при пуше в GitHub  
✅ **Без сложного Kubernetes** - только Docker  
✅ **Простая настройка** на любом сервере  
✅ **Health checks** и мониторинг  
✅ **Rollback** при проблемах  

## 🚀 **Быстрый старт (5 минут)**

### 1. **Настройка GitHub Secrets**

В вашем GitHub репозитории:
1. Перейдите в **Settings** → **Secrets and variables** → **Actions**
2. Добавьте секреты:

```
SERVER_HOST=192.168.1.100     # IP вашего сервера
SERVER_USER=ubuntu             # Пользователь на сервере
SERVER_SSH_KEY=-----BEGIN...   # SSH приватный ключ
SERVER_PORT=22                 # SSH порт (обычно 22)
```

### 2. **Подготовка сервера**

```bash
# Установить Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
sudo systemctl enable docker
sudo systemctl start docker

# Создать SSH ключ для GitHub Actions
ssh-keygen -t rsa -b 4096 -C "github-actions"
# Скопировать публичный ключ в ~/.ssh/authorized_keys
cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys

# Скопировать приватный ключ в GitHub Secrets
cat ~/.ssh/id_rsa
```

### 3. **Первый деплой**

Просто сделайте **push в main ветку** - GitHub Actions автоматически:
1. Соберет Docker образ
2. Запустит тесты
3. Задеплоит на сервер
4. Проверит health

## 🔧 **Ручной деплой (без GitHub Actions)**

### Локально:
```bash
# Собрать и запустить
./scripts/deploy-server.sh build deploy

# Проверить статус
./scripts/deploy-server.sh status

# Посмотреть логи
./scripts/deploy-server.sh logs
```

### На сервере:
```bash
# Скачать код
git clone <your-repo>
cd it-camp

# Собрать и деплоить
./scripts/deploy-server.sh build deploy

# Проверить
curl http://localhost/health
```

## 📋 **Что происходит при деплое:**

1. **Остановка** старого контейнера
2. **Скачивание** нового образа
3. **Запуск** нового контейнера с:
   - Автоперезапуск при падении
   - Health checks каждые 30 секунд
   - Логирование
   - Переменные окружения
4. **Проверка** готовности сервиса

## 🌐 **Доступ к сервису:**

- **Локально на сервере**: `http://localhost/`
- **Извне**: `http://YOUR_SERVER_IP/`
- **Health check**: `http://YOUR_SERVER_IP/health`

## 📊 **Мониторинг:**

```bash
# Статус контейнера
./scripts/deploy-server.sh status

# Логи в реальном времени
./scripts/deploy-server.sh logs

# Статистика ресурсов
docker stats remarks

# Health check
curl -f http://YOUR_SERVER_IP/health
```

## 🔄 **Обновление:**

### Автоматически:
- Просто сделайте `git push origin main`
- GitHub Actions автоматически обновит сервис

### Вручную:
```bash
# Остановить
./scripts/deploy-server.sh stop

# Обновить код
git pull origin main

# Запустить заново
./scripts/deploy-server.sh deploy
```

## 🆘 **Troubleshooting:**

### Сервис не запускается:
```bash
# Проверить логи
./scripts/deploy-server.sh logs

# Проверить статус
./scripts/deploy-server.sh status

# Перезапустить
./scripts/deploy-server.sh restart
```

### Порт занят:
```bash
# Найти что использует порт 80
sudo netstat -tlnp | grep :80

# Остановить конфликтующий сервис
sudo systemctl stop nginx  # если nginx
```

### Docker проблемы:
```bash
# Проверить Docker
docker info

# Перезапустить Docker
sudo systemctl restart docker

# Очистить все
./scripts/deploy-server.sh cleanup
```

## 🔒 **Безопасность:**

### Firewall:
```bash
# Открыть только нужные порты
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw enable
```

### SSH:
```bash
# Отключить парольную аутентификацию
sudo nano /etc/ssh/sshd_config
# PasswordAuthentication no
sudo systemctl restart sshd
```

## 📈 **Масштабирование:**

### Несколько экземпляров:
```bash
# Запустить на разных портах
docker run -d --name remarks-1 -p 8081:8080 your-image
docker run -d --name remarks-2 -p 8082:8080 your-image
```

### Load balancer:
```bash
# Установить nginx для балансировки
sudo apt install nginx
# Настроить upstream в /etc/nginx/sites-available/default
```

## 🎉 **Готово!**

Теперь у вас есть:
- ✅ Автоматический деплой при каждом push
- ✅ Простой мониторинг и логи
- ✅ Автоперезапуск при сбоях
- ✅ Health checks
- ✅ Простота управления

**Никакого Kubernetes!** Только Docker + GitHub Actions + ваш сервер.

## 📞 **Поддержка:**

При проблемах:
1. Проверьте логи: `./scripts/deploy-server.sh logs`
2. Проверьте статус: `./scripts/deploy-server.sh status`
3. Создайте Issue в GitHub
4. Обратитесь к команде DevOps
