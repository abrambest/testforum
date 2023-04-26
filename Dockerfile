FROM golang:latest

WORKDIR /app

COPY . .

RUN go mod download

COPY . .

RUN go build ./cmd/web/

CMD [ "./web" ]