# loadbalancer-provider-haproxy

The loadbalancer-provider-haproxy is the provider for haproxy use-cases. On a high-level it receives messages from the load-balancer-api to create/delete loadbalancers by getting an ip from the ipam-api, and then publishes another message to the operator to finally assign the ip to the loadbalancer.

## Development and Contributing

- [Development Guide](docs/development.md)
- [Contributing](https://infratographer.com/community/contributing/)

## Code of Conduct

[Contributor Code of Conduct](https://infratographer.com/community/code-of-conduct/). By participating in this project you agree to abide by its terms.

## Contact

To contact the maintainers, please open a [GithHub Issue](https://github.com/infratographer/loadbalancer-provider-haproxy/issues/new)

## License

[Apache License, Version 2.0](LICENSE)
