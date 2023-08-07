FROM golang:latest

WORKDIR /app

COPY . .

RUN go build -o benchmarker

CMD ["./benchmarker", "./query_params.csv"]

