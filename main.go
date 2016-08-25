package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/colorstring"
)

const (
	// BANNER is what is printed for help/info output.
	BANNER = "certok - %s\n"

	// VERSION is the binary version.
	VERSION = "v0.1.0"

	defaultWarningDays = 30
)

var (
	days   int
	months int
	years  int

	all bool

	debug   bool
	version bool
)

func init() {
	// parse flags
	flag.IntVar(&years, "years", 0, "Warn if the certificate will expire within this many years.")
	flag.IntVar(&months, "months", 0, "Warn if the certificate will expire within this many months.")
	flag.IntVar(&days, "days", 0, "Warn if the certificate will expire within this many days.")

	flag.BoolVar(&all, "all", false, "Show entire certificate chain, not just the first.")

	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.BoolVar(&version, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "d", false, "run in debug mode")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(BANNER, VERSION))
		flag.PrintDefaults()
	}

	flag.Parse()

	if version {
		fmt.Printf("%s", VERSION)
		os.Exit(0)
	}

	// set log level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// set the default warning days if not set already
	if years == 0 && months == 0 && days == 0 {
		days = defaultWarningDays
	}

}

func main() {
	args := flag.Args()

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

	// create the writer
	w := tabwriter.NewWriter(os.Stdout, 20, 1, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSUBJECT\tISSUER\tALGO\tEXPIRES\tSUNSET DATE")

	// create the WaitGroup
	var wg sync.WaitGroup
	for scanner.Scan() {
		wg.Add(1)
		h := scanner.Text()
		go func() {
			certs, err := checkHost(h, twarn)
			if err != nil {
				logrus.Warn(err)
			}
			for _, cert := range certs {
				sunset := ""
				if cert.sunset != nil {
					sunset = cert.sunset.date.Format("Jan 02, 2006")

				}
				expires := cert.expires
				if cert.warn {
					expires = colorstring.Color("[red]" + cert.expires + "[reset]")
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", cert.name, cert.subject, cert.issuer, cert.algo, expires, sunset)
			}
			wg.Done()
		}()
	}
	// wait for all the goroutines to finish
	wg.Wait()
	// flush the writer
	w.Flush()
}

type host struct {
	name    string
	subject string
	algo    string
	issuer  string
	expires string
	warn    bool
	sunset  *sunsetSignatureAlgorithm
}

func checkHost(h string, twarn time.Time) (map[string]host, error) {
	if !strings.Contains(h, ":") {
		// default to 443
		h += ":443"
	}
	c, err := tls.Dial("tcp", h, nil)
	if err != nil {
		return nil, fmt.Errorf("tcp dial %s failed: %v", h, err)
	}
	defer c.Close()

	certs := make(map[string]host)
	for _, chain := range c.ConnectionState().VerifiedChains {
		for n, cert := range chain {
			if _, checked := certs[string(cert.Signature)]; checked {
				continue
			}
			if !all && n >= 1 {
				continue
			}

			host := host{
				name:    h,
				subject: cert.Subject.CommonName,
				issuer:  cert.Issuer.CommonName,
				algo:    cert.SignatureAlgorithm.String(),
			}

			// check the expiration
			if twarn.After(cert.NotAfter) {
				host.warn = true
			}
			expiresIn := int64(cert.NotAfter.Sub(time.Now()).Hours())
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

			certs[string(cert.Signature)] = host
		}
	}

	return certs, nil
}

func usageAndExit(message string, exitCode int) {
	if message != "" {
		fmt.Fprintf(os.Stderr, message)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(exitCode)
}
