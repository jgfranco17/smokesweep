# BUILD STAGE

FROM golang:1.21 AS build-stage

COPY . /app
WORKDIR /app
RUN go mod download all \
    && CGO_ENABLED=0 GOOS=linux go build -o ./smokesweep ./main.go

# RELEASE STAGE

FROM alpine:3.18 AS release-stage

RUN apk update && apk --no-cache add ca-certificates=20240226-r0
COPY --from=build-stage /app/smokesweep /smokesweep

CMD ["/backend", "--port=8080", "--dev=false"]
