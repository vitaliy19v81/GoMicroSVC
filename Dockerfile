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

# Копируем скрипт create_topic.sh из builder-этапа
#COPY --from=builder /go_microsvc/create_topic.sh .

# Копируем скрипт wait-for-it.sh в контейнер
#COPY wait-for-it.sh /usr/local/bin/wait-for-it.sh

# Добавляем права на выполнение для бинарных файлов и скриптов
#RUN chmod +x /usr/local/bin/wait-for-it.sh ./create_topic.sh ./main

# Экспортируем порт
EXPOSE 8080

# Добавляем путь к kafka-topics.sh в PATH
#ENV PATH="/opt/kafka/bin:${PATH}"

# Запуск приложения и скрипта
CMD ["./main"]



