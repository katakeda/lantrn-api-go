FROM golang:latest

WORKDIR /app
COPY . .

RUN go install
RUN make clean && make build

CMD ["./lantrn-api-go"]