FROM golang:1.20
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go get github.com/hoyle1974/grapevine/client@latest
RUN go get github.com/hoyle1974/grapevine/services@latest
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /cmd
CMD ["/cmd"]
