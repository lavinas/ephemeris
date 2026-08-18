FROM golang:1.26-alpine

WORKDIR /app

RUN apk add --no-cache git gcc musl-dev chromium fontconfig ttf-freefont

RUN go install github.com/air-verse/air@latest

ENV PATH="/go/bin:${PATH}"

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8083

CMD ["sh", "-c", "sleep 5 && air"]



