FROM gcr.io/distroless/static

# Copy the binary that goreleaser built
COPY  loadbalancer-provider-haproxy /loadbalancer-provider-haproxy

# Run the web service on container startup.
ENTRYPOINT ["/loadbalancer-provider-haproxy"]
CMD ["process"]
