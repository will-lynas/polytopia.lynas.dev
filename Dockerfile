FROM golang:1.18.0

WORKDIR /usr/src/app

COPY go.mod ./
RUN go mod download

COPY . .

EXPOSE 8443

RUN go build -o go-server
CMD ["./go-server"]
