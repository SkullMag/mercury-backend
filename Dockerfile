FROM golang:1.18

WORKDIR /usr/src/app

ENV PORT="9090" 

ADD https://github.com/ufoscout/docker-compose-wait/releases/download/2.9.0/wait /wait
RUN chmod +x /wait

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . . 

RUN go build -v -o /usr/src/app/executable
EXPOSE 9090 

CMD /wait && /usr/src/app/executable
