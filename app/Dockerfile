# build stage
FROM golang:1.22 AS builder
WORKDIR /src

# 1) сначала модули (чтобы кэшировалось)
COPY go.mod ./
# копируем go.sum, если он уже есть; если нет — не страшно
# (оставь строку, Docker проигнорирует, если файла нет)
COPY go.sum ./
RUN go mod download

# 2) теперь весь исходник
COPY . .

# на случай отсутствия go.sum — сформируем его
RUN go mod tidy

# 3) сборка
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/app .

# runtime stage
FROM gcr.io/distroless/base-debian12
ENV TZ=Etc/UTC
WORKDIR /
COPY --from=builder /bin/app /app
EXPOSE 8081
ENTRYPOINT ["/app"]
