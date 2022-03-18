FROM golang:1.18

WORKDIR /usr/src/app

ENV PORT="9090" 

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . . 

RUN go build -v -o /usr/src/app/executable
EXPOSE 9090 

CMD ["/usr/src/app/executable"]
