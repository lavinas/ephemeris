FROM golang:1.26-alpine

WORKDIR /app

RUN apk add --no-cache git gcc musl-dev

RUN go install github.com/air-verse/air@latest

ENV PATH="/go/bin:${PATH}"

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["air"]



