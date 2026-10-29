FROM golang
LABEL authors="sharifi"

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ./app/server ./cmd/httpserver

EXPOSE 8080

ENTRYPOINT ["./app/server"]