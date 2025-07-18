FROM golang:1.24

## We create an /app directory within our
## image that will hold our application source
## files
RUN mkdir /app

## We specify that we now wish to execute 
## any further commands inside our /app
## directory
WORKDIR /app

COPY go.mod /app
COPY go.sum /app
RUN go mod tidy

## We copy everything in the root directory
## into our /app directory
ADD . /app

## we run go build to compile the binary
## executable of our Go program
RUN go build -o main .


## Our start command which kicks off
## our newly created binary executable
CMD ["/app/main"]