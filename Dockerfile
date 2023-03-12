FROM golang:latest

WORKDIR /app
COPY . .

RUN go install -buildvcs=false
RUN make clean && make build

CMD ["./lantrn-api-go"]