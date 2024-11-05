# Temel Go imajını kullan
FROM golang:1.22 AS builder

# Çalışma dizinini ayarla
WORKDIR /app

# Go mod ve sum dosyalarını kopyala
COPY go.mod go.sum ./

# Bağımlılıkları yükle
RUN go mod download

# Uygulama kodunu kopyala
COPY . .

# Uygulamayı derle
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# İkinci aşama: minimal bir imajda uygulamayı çalıştır
FROM alpine:latest

# Gerekli paketleri yükle
RUN apk --no-cache add ca-certificates

# Uygulama dosyasını kopyala
WORKDIR /root/
COPY --from=builder /app/main .

# Uygulamayı çalıştır
CMD ["./main"]
