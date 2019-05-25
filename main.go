package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/genuinetools/certok/version"
	"github.com/genuinetools/pkg/cli"
	"github.com/mitchellh/colorstring"
	"github.com/sirupsen/logrus"
)

//

const (
	defaultWarningDays = 30
)

var (
	days   int
	months int
	years  int

	all bool

	debug bool
)

func main() {
	// Create a new cli program.
	p := cli.NewProgram()
	p.Name = "certok"
	p.Description = "A tool to check the validity and expiration dates of SSL certificates"

	// Set the GitCommit and Version.
	p.GitCommit = version.GITCOMMIT
	p.Version = version.VERSION

	// Setup the global flags.
	p.FlagSet = flag.NewFlagSet("global", flag.ExitOnError)
	p.FlagSet.IntVar(&years, "years", 0, "Warn if the certificate will expire within this many years.")
	p.FlagSet.IntVar(&months, "months", 0, "Warn if the certificate will expire within this many months.")
	p.FlagSet.IntVar(&days, "days", 0, "Warn if the certificate will expire within this many days.")

	p.FlagSet.BoolVar(&all, "all", false, "Show entire certificate chain, not just the first.")

	p.FlagSet.BoolVar(&debug, "d", false, "enable debug logging")

	// Set the before function.
	p.Before = func(ctx context.Context) error {
		// Set the log level.
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		// set the default warning days if not set already
		if years == 0 && months == 0 && days == 0 {
			days = defaultWarningDays
		}

		return nil
	}

	// Set the main program action.
	p.Action = func(ctx context.Context, args []string) error {
		// check if we are reading from a file or stdin
		var (
			scanner *bufio.Scanner
		)
		if len(args) == 0 {
			logrus.Debugf("no file passed, reading from stdin...")
			scanner = bufio.NewScanner(os.Stdin)
		} else {
			f, err := os.Open(args[0])
			if err != nil {
				logrus.Fatalf("opening file %s failed: %v", args[0], err)
				os.Exit(1)
			}
			defer f.Close()
			scanner = bufio.NewScanner(f)
		}

		// get the time now
		now := time.Now()
		twarn := now.AddDate(years, months, days)

		// create the WaitGroup
		var wg sync.WaitGroup
		hosts := hosts{}

		for scanner.Scan() {
			wg.Add(1)
			h := scanner.Text()
			go func() {
				certs, err := checkHost(h, twarn)
				if err != nil {
					logrus.Warn(err)
				}
				hosts = append(hosts, host{name: h, certs: certs})
				wg.Done()
			}()
		}

		// wait for all the goroutines to finish
		wg.Wait()

		// Sort the hosts
		sort.Sort(hosts)

		// create the writer
		w := tabwriter.NewWriter(os.Stdout, 20, 1, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSUBJECT\tISSUER\tALGO\tEXPIRES\tSUNSET DATE\tERROR")

		// Iterate over the hosts
		for i := 0; i < len(hosts); i++ {
			for _, cert := range hosts[i].certs {
				sunset := ""
				if cert.sunset != nil {
					sunset = cert.sunset.date.Format("Jan 02, 2006")

				}
				expires := cert.expires
				if cert.warn {
					expires = colorstring.Color("[red]" + cert.expires + "[reset]")
				}
				error := cert.error
				if error != "" {
					error = colorstring.Color("[red]" + cert.error + "[reset]")
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", cert.name, cert.subject, cert.issuer, cert.algo, expires, sunset, error)
			}
		}

		// flush the writer
		w.Flush()

		return nil
	}

	// Run our program.
	p.Run()
}

type hosts []host

func (h hosts) Len() int           { return len(h) }
func (h hosts) Less(i, j int) bool { return h[i].name < h[j].name }
func (h hosts) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

type host struct {
	name  string
	certs map[string]certificate
}

type certificate struct {
	name    string
	subject string
	algo    string
	issuer  string
	expires string
	warn    bool
	error   string
	sunset  *sunsetSignatureAlgorithm
}

func checkHost(h string, twarn time.Time) (map[string]certificate, error) {
	if !strings.Contains(h, ":") {
		// default to 443
		h += ":443"
	}
	c, err := tls.Dial("tcp", h, nil)
	if err != nil {
		switch cerr := err.(type) {
		case x509.CertificateInvalidError:
			ht := createHost(h, twarn, cerr.Cert)
			ht.error = err.Error()
			return map[string]certificate{
				string(cerr.Cert.Signature): ht,
			}, nil
		case x509.UnknownAuthorityError:
			ht := createHost(h, twarn, cerr.Cert)
			ht.error = err.Error()
			return map[string]certificate{
				string(cerr.Cert.Signature): ht,
			}, nil
		case x509.HostnameError:
			ht := createHost(h, twarn, cerr.Certificate)
			ht.error = err.Error()
			return map[string]certificate{
				string(cerr.Certificate.Signature): ht,
			}, nil
		}
		return nil, fmt.Errorf("tcp dial %s failed: %v", h, err)
	}
	defer c.Close()

	certs := make(map[string]certificate)
	for _, chain := range c.ConnectionState().VerifiedChains {
		for n, cert := range chain {
			if _, checked := certs[string(cert.Signature)]; checked {
				continue
			}
			if !all && n >= 1 {
				continue
			}

			ht := createHost(h, twarn, cert)

			certs[string(cert.Signature)] = ht
		}
	}

	return certs, nil
}

func createHost(name string, twarn time.Time, cert *x509.Certificate) certificate {
	host := certificate{
		name:    name,
		subject: cert.Subject.CommonName,
		issuer:  cert.Issuer.CommonName,
		algo:    cert.SignatureAlgorithm.String(),
	}

	// check the expiration
	if twarn.After(cert.NotAfter) {
		host.warn = true
	}
	expiresIn := int64(time.Until(cert.NotAfter).Hours())
	if expiresIn <= 48 {
		host.expires = fmt.Sprintf("%d hours", expiresIn)
	} else {
		host.expires = fmt.Sprintf("%d days", expiresIn/24)
	}

	// Check the signature algorithm, ignoring the root certificate.
	if alg, exists := sunsetSignatureAlgorithms[cert.SignatureAlgorithm]; exists {
		if cert.NotAfter.Equal(alg.date) || cert.NotAfter.After(alg.date) {
			host.warn = true
		}
		host.sunset = &alg
	}

	return host
}
