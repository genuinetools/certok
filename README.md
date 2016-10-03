# certok

[![Travis CI](https://travis-ci.org/jessfraz/certok.svg?branch=master)](https://travis-ci.org/jessfraz/certok)

Command line tool to check the validity and expiration dates of SSL certificates.

## Usage

Pass a line deliminated file filled with hostnames to either stdin or the first
argument of the command. For example:

```console
$ certok ~/hostsfile
NAME                 SUBJECT                     ISSUER                                           ALGO                EXPIRES             SUNSET DATE
telize.j3ss.co:443   telize.j3ss.co              Let's Encrypt Authority X3                       SHA256-RSA          77 days
r.j3ss.co:443        r.j3ss.co                   Let's Encrypt Authority X3                       SHA256-RSA          77 days
contained.af:443     contained.af                Let's Encrypt Authority X3                       SHA256-RSA          77 days
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
