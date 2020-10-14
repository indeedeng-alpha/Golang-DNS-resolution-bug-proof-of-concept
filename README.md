![OSS Lifecycle](https://img.shields.io/osslifecycle/indeedeng-alpha/Golang-DNS-resolution-bug-proof-of-concept.svg)

This project is a proof-of-concept which demonstrates [incorrect behavior](https://github.com/golang/go/issues/23427) with Golang's DNS resolver & connection re-use with
the http client.

If you execute `run.sh`, then following happens:
1. two interfaces eth10 and eth11 on 10.0.0.1 and 10.0.0.2 are created respectively
2. go run's the dns.go main
3. go run's the dns.go main with --close-idle-conns flag set
4. cleans up the two interfaces

If you execute and see the message `failed test` in the first half of the output, the bug occurred.

The bug demonstrated in the dns.go is the following:

A DNS server is started on :8053 and it contains a single A record for `foo.service.` in `10.0.0.1`.
The default go resolver is switched on and the resolver dialer is set to point to the `:8053` dns server.

Then an HTTP server is started on `10.0.0.1:8080`, and a second HTTP server is started on `10.0.0.2.:8080`.

Finally, an HTTP request is made to the A address, `foo.service:8080` and the contents of the response is printed.
The A record of `foo.service.` is updated to point to `10.0.0.2`. Then an HTTP request is made again to the A address,
`foo.service:8080`. If Go does the right thing, it will hit the HTTP server on `10.0.0.2:8080` and exit 0. If Go does
the wrong thing, it will hit the first HTTP server on `10.0.0.1:8080` and exit 1.

## Project Maintainers
- @cnmcavoy

## Code of Conduct
This project is governed by the [Contributor Covenant v 1.4.1](CODE_OF_CONDUCT.md).

## License
This project uses the [Apache 2.0](LICENSE) license.
