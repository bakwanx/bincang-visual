FROM golang:1.24

## We specify that we now wish to execute 
## any further commands inside our /app
## directory
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

CMD ["./main"]