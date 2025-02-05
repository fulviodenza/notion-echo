FROM golang:1.23 as builder

# Create and change to the app directory.
WORKDIR /app
ADD . /app

RUN go mod tidy && go mod vendor
RUN go build -o /notion-echo
COPY run.sh /run.sh
RUN chmod +x /run.sh
CMD [ "./run.sh" ]
