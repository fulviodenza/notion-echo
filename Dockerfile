FROM golang:1.21 as builder

# Create and change to the app directory.
WORKDIR /app
ADD . /app

RUN go build -o /notion-echo

CMD [ "/notion-echo" ]
