FROM golang:1.21.3

WORKDIR /app

COPY . .

RUN go build -o myapp

EXPOSE 9090

CMD ["./myapp"]