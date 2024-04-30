FROM golang:latest

RUN apt-get update && apt-get install -y iperf iputils-ping

WORKDIR /app

COPY . .

RUN go mod tidy && go build -o exporter .

CMD ["./exporter"]

