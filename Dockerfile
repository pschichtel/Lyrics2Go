FROM golang:latest AS build
WORKDIR /go/src/Lyrics2Go
COPY *.go ./
RUN go get .
ENV CGO_ENABLED=0 \
    GOOS=linux
RUN go build -o lyrics2go -a -ldflags '-extldflags "-static"' .

FROM alpine:latest
RUN apk apk --update upgrade
RUN apk add ca-certificates
COPY --from=build /go/src/Lyrics2Go/lyrics2go /bin/lyrics2go
ENTRYPOINT ["/bin/lyrics2go", "/provider"]
