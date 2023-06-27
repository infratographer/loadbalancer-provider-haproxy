FROM gcr.io/distroless/static

# Copy the binary that goreleaser built
COPY  loadbalancer-provider-proxy /loadbalancer-provider-proxy

# Run the web service on container startup.
ENTRYPOINT ["/loadbalancer-provider-proxy"]
CMD ["process"]
