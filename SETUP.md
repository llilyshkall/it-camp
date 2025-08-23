# Пошаговая настройка проекта

## 🎯 **Что мы настроим:**

✅ **Автоматический деплой** при пуше в GitHub  
✅ **Без сложного Kubernetes** - только Docker  
✅ **Простая настройка** на любом сервере  
✅ **Health checks** и мониторинг  

## 🚀 **Шаг 1: Подготовка GitHub репозитория**

### 1.1 Создание Personal Access Token
1. Зайдите в GitHub → **Settings** (шестеренка в правом верхнем углу)
2. В левом меню найдите **Developer settings**
3. Нажмите **Personal access tokens** → **Tokens (classic)**
4. Нажмите **Generate new token (classic)**
5. В **Note** напишите: `GHCR Token for IT-Camp`
6. Выберите **Expiration**: `No expiration` (или укажите срок)
7. **ВАЖНО**: Поставьте галочку `write:packages`
8. Нажмите **Generate token**
9. **Скопируйте токен** (он показывается только один раз!)

### 1.2 Обновление Makefile
В файле `Makefile` замените:
```makefile
IMAGE_NAME := ghcr.io/ВАШ_РЕАЛЬНЫЙ_USERNAME/remarks
```

## 🔧 **Шаг 2: Подготовка сервера**

### 2.1 Установка Docker
```bash
# Установить Docker
curl -fsSL https://get.docker.com | sh

# Добавить пользователя в группу docker
sudo usermod -aG docker $USER

# Включить и запустить Docker
sudo systemctl enable docker
sudo systemctl start docker

# Проверить установку
docker --version
docker run hello-world
```

### 2.2 Создание SSH ключа для GitHub Actions
```bash
# Создать SSH ключ
ssh-keygen -t rsa -b 4096 -C "github-actions"

# Добавить публичный ключ в authorized_keys
cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys

# Скопировать приватный ключ (нужен для GitHub Secrets)
cat ~/.ssh/id_rsa
```

## 🔐 **Шаг 3: Настройка GitHub Secrets**

В вашем GitHub репозитории:
1. Перейдите в **Settings** → **Secrets and variables** → **Actions**
2. Нажмите **New repository secret**
3. Добавьте секреты:

```
SERVER_HOST=192.168.1.100     # IP вашего сервера
SERVER_USER=ubuntu             # Пользователь на сервере
SERVER_SSH_KEY=-----BEGIN...   # Приватный SSH ключ (весь блок)
SERVER_PORT=22                 # SSH порт (обычно 22)
```

## 🚀 **Шаг 4: Первый деплой**

### 4.1 Автоматический деплой
Просто сделайте **push в main ветку**:
```bash
git add .
git commit -m "Initial setup"
git push origin main
```

GitHub Actions автоматически:
1. Соберет Docker образ
2. Запустит тесты
3. Задеплоит на сервер
4. Проверит health

### 4.2 Ручной деплой (если нужно)
```bash
# На сервере
git clone <your-repo-url>
cd it-camp

# Собрать и запустить
./scripts/deploy-server.sh build deploy

# Проверить статус
./scripts/deploy-server.sh status
```

## 🌐 **Шаг 5: Проверка работы**

### 5.1 Проверка на сервере
```bash
# Проверить статус контейнера
./scripts/deploy-server.sh status

# Посмотреть логи
./scripts/deploy-server.sh logs

# Health check
curl http://localhost/health
```

### 5.2 Проверка извне
```bash
# С вашего компьютера
curl http://YOUR_SERVER_IP/health
```

## 🔒 **Шаг 6: Настройка безопасности**

### 6.1 Firewall
```bash
# Открыть только нужные порты
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw enable

# Проверить статус
sudo ufw status
```

### 6.2 SSH безопасность
```bash
# Отключить парольную аутентификацию
sudo nano /etc/ssh/sshd_config

# Найти и изменить:
# PasswordAuthentication no
# PermitRootLogin no

# Перезапустить SSH
sudo systemctl restart sshd
```

## 📊 **Шаг 7: Мониторинг**

### 7.1 Основные команды
```bash
# Статус сервиса
./scripts/deploy-server.sh status

# Логи в реальном времени
./scripts/deploy-server.sh logs

# Статистика ресурсов
docker stats remarks

# Health check
curl -f http://YOUR_SERVER_IP/health
```

### 7.2 Автоматический мониторинг
GitHub Actions автоматически проверяет health после каждого деплоя.

## 🔄 **Шаг 8: Обновление**

### 8.1 Автоматическое обновление
Просто делайте `git push origin main` - GitHub Actions автоматически обновит сервис.

### 8.2 Ручное обновление
```bash
# На сервере
cd it-camp
git pull origin main

# Перезапустить сервис
./scripts/deploy-server.sh deploy
```

## 🆘 **Troubleshooting**

### Проблемы с деплоем
```bash
# Проверить статус
./scripts/deploy-server.sh status

# Посмотреть логи
./scripts/deploy-server.sh logs

# Перезапустить
./scripts/deploy-server.sh restart
```

### Проблемы с Docker
```bash
# Проверить Docker
docker info

# Перезапустить Docker
sudo systemctl restart docker

# Очистить все
./scripts/deploy-server.sh cleanup
```

### Проблемы с сетью
```bash
# Проверить порты
sudo netstat -tlnp | grep :80

# Проверить firewall
sudo ufw status
```

## ✅ **Проверочный список**

- [ ] Personal Access Token создан с `write:packages` scope
- [ ] Makefile обновлен с вашим username
- [ ] Docker установлен на сервере
- [ ] SSH ключ создан и добавлен в authorized_keys
- [ ] GitHub Secrets настроены
- [ ] Первый деплой прошел успешно
- [ ] Сервис доступен по IP
- [ ] Health check работает
- [ ] Firewall настроен
- [ ] SSH безопасность настроена

## 🎉 **Готово!**

Теперь у вас есть:
- ✅ Автоматический деплой при каждом push
- ✅ Простой мониторинг и логи
- ✅ Автоперезапуск при сбоях
- ✅ Health checks
- ✅ Простота управления

**Никакого Kubernetes!** Только Docker + GitHub Actions + ваш сервер.

## 📞 **Поддержка**

При возникновении проблем:
1. Проверьте логи: `./scripts/deploy-server.sh logs`
2. Проверьте статус: `./scripts/deploy-server.sh status`
3. Создайте Issue в GitHub
4. Обратитесь к команде DevOps
