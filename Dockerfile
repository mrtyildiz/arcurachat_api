# 1. Go temel imajı
FROM golang:1.20

# 2. Çalışma dizinini ayarla
WORKDIR /arcurachat_api

# 3. Bağımlılıkları kopyala ve yükle
#COPY go.mod go.sum ./
#RUN go mod download

# 4. Uygulama dosyalarını kopyala
COPY . .
RUN go mod init arcurachat_api
RUN go mod tidy

# 5. Uygulamayı derle
#RUN go build -o gin-auth

# 6. API'yi başlat
#CMD ["/app/gin-auth"]
