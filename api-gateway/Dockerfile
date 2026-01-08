FROM golang:1.24-alpine AS build
WORKDIR /src
COPY . .
RUN go build -o api .

FROM alpine:3.20
WORKDIR /app
COPY --from=build /src/api .
EXPOSE 8080
CMD ["./api"]
