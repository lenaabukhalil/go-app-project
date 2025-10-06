# ===== Build stage =====
FROM golang:1.22-alpine AS builder
WORKDIR /src
# ننسخ ملف السورس الوحيد
COPY main.go .
# نولّد/نحدّث الموديولات ببساطة (بدون go.mod مسبق)
RUN go mod init lina/goapp || true \
    && go get github.com/go-sql-driver/mysql@v1.7.1 \
    && go mod tidy
# نبني باينري ستاتيك وخفيف
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server main.go

# ===== Runtime stage (صغير وآمن) =====
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /src/server /app/server
EXPOSE 8000
USER 65532:65532
ENTRYPOINT ["/app/server"]
