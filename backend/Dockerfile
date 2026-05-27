## Dockerfile

FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . .
RUN go mod tidy && go mod download && go test -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -o /todoapp .

FROM alpine:latest
COPY --from=build /todoapp /todoapp
ENV Storage=/data/todos.db
ENV listenPort=8080
RUN mkdir -p /data
EXPOSE 8080
HEALTHCHECK CMD wget -qO- http://localhost:8080/health || exit 1
ENTRYPOINT ["/todoapp"]
