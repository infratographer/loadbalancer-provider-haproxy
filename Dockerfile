FROM golang:1.20 as build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -o /go/bin/loadbalancer-provider-haproxy

FROM gcr.io/distroless/static

# Copy the binary that goreleaser built
COPY --from=build /go/bin/loadbalancer-provider-haproxy /loadbalancer-provider-haproxy

# Run the web service on container startup.
ENTRYPOINT ["/loadbalancer-provider-haproxy"]
CMD ["process"]
