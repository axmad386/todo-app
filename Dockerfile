# lightweight container for go
FROM golang:alpine AS build

# update container's packages and install git
RUN apk update && apk add --no-cache git

# set /todo to be the active directory
WORKDIR /todo

# copy all files from outside container, into the container
COPY . .

# build bin
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w' -o bin *.go

FROM alpine:latest

WORKDIR /todo

COPY --from=build /todo/bin /todo/bin

EXPOSE 3030

CMD [ "/todo/bin" ]
