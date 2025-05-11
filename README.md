# finalproject
**Распределённый калькулятор** с поддержкой:

- Регистрации и JWT-авторизации  
- Хранения данных в SQLite (через GORM)  
- Agent’а, который опрашивает задачи по HTTP и возвращает результаты  
- Веб-интерфейса на React + Vite + TailwindCSS  
- End-to-end и unit-тестов


## Требования

- **Go** ≥ 1.20 (рекомендуется 1.23+)  
- **Node.js** ≥ 16 и **npm** или **yarn**  

---

## Установка и запуск

### Backend без фронтенда

1. Клонируйте репозиторий и перейдите в папку проекта:
   ```bash
   git clone https://github.com/egocentri/finalproject.git
   cd finalproject
   ```

2. Запустите HTTP-сервис (оркестратор):
   ```bash
   go run ./cmd/orchestrator/...
   ```
   - Слушает на `http://localhost:8080`  
   - Автоматически создает файл `data.db`

### Agent

В отдельном терминале **из того же корня**:

```bash
go run ./cmd/agent/...
```

- Agent будет опрашивать `GET http://localhost:8080/api/v1/internal/task`  
- Симулировать задержку и отправлять результат на `POST http://localhost:8080/api/v1/internal/task`

### Frontend

1. Перейдите в директорию фронтенда:
   ```bash
   cd frontend
   ```

2. Установите зависимости:
   ```bash
   npm install
   # или, если вы используете yarn:
   # yarn install
   ```

3. Запустите dev-сервер:
   ```bash
   npm run dev
   # или yarn dev
   ```
   - Dev-сервер Vite по умолчанию на `http://localhost:5173`  
   - Он проксирует все запросы `/api/v1/**` → `http://localhost:8080`

4. Откройте в браузере `http://localhost:5173` и работайте через веб-интерфейс.

---

## 🔧 Использование API без фронтенда

Можно оперировать cURL / Postman:

```bash
# 1. Регистрация
curl -i -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"login":"user1","password":"pass1"}'

# 2. Логин → получаем JSON { "token": "..." }
curl -s -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"login":"user1","password":"pass1"}'

# 3. Вычислить выражение (Bearer токен)
TOKEN=eyJhbGciOi...
curl -i -X POST http://localhost:8080/api/v1/calculate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"expression":"(2+3)*4/2"}'

# 4. Получить историю
curl -s -X GET http://localhost:8080/api/v1/expressions \
  -H "Authorization: Bearer $TOKEN"
```

---

## 🛠 Исправление ошибки `Cannot find module '@vitejs/plugin-react'`

Если при `npm run dev` видите:

```
Error: Cannot find module '@vitejs/plugin-react'
```

нужно:

1. Перейти в папку фронтенда:
   ```bash
   cd frontend
   ```

2. Установить плагин:
   ```bash
   npm install --save-dev @vitejs/plugin-react
   ```

3. Перезапустить:
   ```bash
   npm run dev
   ```

Также убедитесь, что в вашем `package.json` присутствуют:

```json
{
  "devDependencies": {
    "@vitejs/plugin-react": "^4.0.0",
    "vite": "^5.0.0",
    "tailwindcss": "^3.4.4",
    "postcss": "^8.4.24",
    "autoprefixer": "^10.4.14"
  },
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-router-dom": "^6.14.1"
  }
}
```

и затем выполнить `npm install`.

---

## Запуск тестов

### Unit-тесты (Go)

```bash
go test ./tests/unit/...
```

### Integration-тесты (Go)

```bash
go test ./tests/integration/...
```

---

## Полезные команды

```bash
# Общая очистка модуля
go clean -modcache

# Обновление зависимостей
go mod tidy


```

---
