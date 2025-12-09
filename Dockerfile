FROM golang:1.25.5-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy
COPY . .
RUN go build -o app .
EXPOSE 8000
CMD ["./app"]