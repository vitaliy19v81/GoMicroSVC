FROM golang:1.22-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /go_microsvc

# Копируем go.mod и go.sum и скачиваем зависимости
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download

# Копируем все файлы проекта
COPY . .

# Сборка приложения
RUN go build -o main ./cmd/api/main.go

# Создаем минимальный образ
FROM alpine:latest

# Устанавливаем рабочую директорию
WORKDIR /root/

# Копируем скомпилированный бинарник
COPY --from=builder /go_microsvc/main .

# Копируем .env файл
COPY --from=builder /go_microsvc/.env .

# Экспортируем порт
EXPOSE 8080

# Запуск приложения и скрипта
CMD ["./main"]



