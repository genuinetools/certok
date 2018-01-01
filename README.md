# certok

[![Travis CI](https://travis-ci.org/jessfraz/certok.svg?branch=master)](https://travis-ci.org/jessfraz/certok)

Command line tool to check the validity and expiration dates of SSL certificates.

## Installation

#### Binaries

- **darwin** [386](https://github.com/jessfraz/certok/releases/download/v0.1.0/certok-darwin-386) / [amd64](https://github.com/jessfraz/certok/releases/download/v0.1.0/certok-darwin-amd64)
- **freebsd** [386](https://github.com/jessfraz/certok/releases/download/v0.1.0/certok-freebsd-386) / [amd64](https://github.com/jessfraz/certok/releases/download/v0.1.0/certok-freebsd-amd64)
- **linux** [386](https://github.com/jessfraz/certok/releases/download/v0.1.0/certok-linux-386) / [amd64](https://github.com/jessfraz/certok/releases/download/v0.1.0/certok-linux-amd64) / [arm](https://github.com/jessfraz/certok/releases/download/v0.1.0/certok-linux-arm) / [arm64](https://github.com/jessfraz/certok/releases/download/v0.1.0/certok-linux-arm64)
- **solaris** [amd64](https://github.com/jessfraz/certok/releases/download/v0.1.0/certok-solaris-amd64)
- **windows** [386](https://github.com/jessfraz/certok/releases/download/v0.1.0/certok-windows-386) / [amd64](https://github.com/jessfraz/certok/releases/download/v0.1.0/certok-windows-amd64)

#### Via Go

```bash
$ go get github.com/jessfraz/certok
```

## Usage

Pass a line deliminated file filled with hostnames to either stdin or the first
argument of the command. For example:

```console
$ certok ~/hostsfile
NAME                 SUBJECT            ISSUER                        ALGO           EXPIRES     SUNSET DATE
telize.j3ss.co:443   telize.j3ss.co     Let's Encrypt Authority X3    SHA256-RSA     77 days
r.j3ss.co:443        r.j3ss.co          Let's Encrypt Authority X3    SHA256-RSA     77 days
contained.af:443     contained.af       Let's Encrypt Authority X3    SHA256-RSA     77 days
```

```console
$ certok -h
certok - v0.1.0
  -all
        Show entire certificate chain, not just the first.
  -d    run in debug mode
  -days int
        Warn if the certificate will expire within this many days.
  -months int
        Warn if the certificate will expire within this many months.
  -v    print version and exit (shorthand)
  -version
        print version and exit
  -years int
        Warn if the certificate will expire within this many years.
```
