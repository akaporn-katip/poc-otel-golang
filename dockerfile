FROM golang:1.24.5-alpine as build

WORKDIR /app
COPY . ./
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o /app/main .



FROM alpine
COPY --from=build /app/main /bin
EXPOSE 3333
CMD ["/bin/main"]