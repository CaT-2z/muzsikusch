FROM golang:alpine AS build

WORKDIR ./muzsikusch
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
COPY src ./src
COPY html ./html
RUN go build ./

FROM alpine

WORKDIR ./muzsikusch
COPY --from=build /go/muzsikusch/html ./html
COPY --from=build /go/muzsikusch/uzsikusch ./
ENTRYPOINT ./muzsikusch