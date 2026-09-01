FROM golang:1.24 as builder

# Create and change to the app directory.
WORKDIR /app
ADD . /app

RUN go mod tidy && go mod vendor

# Test stage: `docker build --target tester` runs the suite using the streamed
# build context, which works on the DinD runner where bind mounts don't.
FROM builder as tester
RUN go test ./...

FROM builder as build
RUN go build -o /notion-echo
COPY run.sh /run.sh
RUN chmod +x /run.sh
CMD [ "./run.sh" ]
