FROM golang:1.20.3 as dependencies

WORKDIR /app

RUN go install github.com/cosmtrek/air@latest

COPY go.mod go.sum ./

RUN go mod download 

FROM dependencies as builder  

COPY . .

RUN make build

CMD ["./tmp/main"]

