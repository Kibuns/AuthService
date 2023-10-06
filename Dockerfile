FROM golang:1.20.8-alpine

WORKDIR /app

ENV GO111MODULE=on
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY . ./

RUN go build -o /AuthService

EXPOSE 10000

CMD [ "/AuthService" ]
