package runner

import (
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/projectdiscovery/fileutil"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/goflags"
)

// Options contains the configuration options for tuning
// the active dns resolving process.
type Options struct {
	Directory          string // Directory is a directory for temporary data
	Domain             string // Domain is the domain to find subdomains
	SubdomainsList     string // SubdomainsList is the file containing list of hosts to resolve
	ResolversFile      string // ResolversFile is the file containing resolvers to use for enumeration
	Wordlist           string // Wordlist is a wordlist to use for enumeration
	MassdnsPath        string // MassdnsPath contains the path to massdns binary
	Output             string // Output is the file to write found subdomains to.
	Json               bool   // Json is the format for making output as ndjson
	Silent             bool   // Silent suppresses any extra text and only writes found host:port to screen
	Version            bool   // Version specifies if we should just show version and exit
	Retries            int    // Retries is the number of retries for dns enumeration
	Verbose            bool   // Verbose flag indicates whether to show verbose output or not
	NoColor            bool   // No-Color disables the colored output
	Threads            int    // Thread controls the number of parallel host to enumerate
	MassdnsRaw         string // MassdnsRaw perform wildcards filtering from an existing massdns output file
	WildcardThreads    int    // WildcardsThreads controls the number of parallel host to check for wildcard
	StrictWildcard     bool   // StrictWildcard flag indicates whether wildcard check has to be performed on each found subdomains
	WildcardOutputFile string // StrictWildcard flag indicates whether wildcard check has to be performed on each found subdomains

	Stdin bool // Stdin specifies whether stdin input was given to the process
}

// ParseOptions parses the command line flags provided by a user
func ParseOptions() *Options {
	options := &Options{}

	flagSet := goflags.NewFlagSet()
	flagSet.SetDescription(`shuffleDNS is a wrapper around massdns written in go that allows you to enumerate valid subdomains using active bruteforce as well as resolve subdomains with wildcard handling and easy input-output support.`)

	createGroup(flagSet, "input", "Input",
		flagSet.StringVar(&options.Domain, "d", "", "Domain to find or resolve subdomains for"),
		flagSet.StringVar(&options.ResolversFile, "r", "", "File containing list of resolvers for enumeration"),
		flagSet.StringVar(&options.Wordlist, "w", "", "File containing words to bruteforce for domain"),
	)

	createGroup(flagSet, "rate-limit", "Rate-Limit",
		flagSet.IntVar(&options.Threads, "t", 10000, "Number of concurrent massdns resolves"),
	)

	createGroup(flagSet, "output", "Output",
		flagSet.StringVar(&options.Output, "o", "", "File to write output to (optional)"),
		flagSet.BoolVar(&options.Json, "json", false, "Make output format as ndjson"),
	)

	createGroup(flagSet, "configs", "Configurations",
		flagSet.BoolVar(&options.StrictWildcard, "strict-wildcard", false, "Perform wildcard check on all found subdomains"),
		flagSet.IntVar(&options.WildcardThreads, "wt", 25, "Number of concurrent wildcard checks"),
		flagSet.StringVar(&options.SubdomainsList, "list", "", "File containing list of subdomains to resolve"),
		flagSet.StringVar(&options.MassdnsPath, "massdns", "", "Path to the massdns binary"),
		flagSet.StringVar(&options.Directory, "directory", "", "Temporary directory for enumeration"),
		flagSet.StringVar(&options.MassdnsRaw, "raw-input", "", "Validate raw full massdns output"),
		flagSet.StringVar(&options.WildcardOutputFile, "wildcard-output-file", "", "Dump wildcard ips to output file"),
	)

	createGroup(flagSet, "Optimizations", "Optimizations",
		flagSet.IntVar(&options.Retries, "retries", 5, "Number of retries for dns enumeration"),
	)

	createGroup(flagSet, "debug", "Debug",
		flagSet.BoolVar(&options.Silent, "silent", false, "Show only subdomains in output"),
		flagSet.BoolVar(&options.Version, "version", false, "Show version of shuffledns"),
		flagSet.BoolVar(&options.Verbose, "v", false, "Show Verbose output"),
		flagSet.BoolVar(&options.NoColor, "nC", false, "Don't Use colors in output"),
	)

	_ = flagSet.Parse()

	// Check if stdin pipe was given
	options.Stdin = fileutil.HasStdin()

	// Read the inputs and configure the logging
	options.configureOutput()

	// Show the user the banner
	showBanner()

	if options.Version {
		gologger.Info().Msgf("Current Version: %s\n", Version)
		os.Exit(0)
	}
	// Validate the options passed by the user and if any
	// invalid options have been used, exit.
	err := options.validateOptions()
	if err != nil {
		gologger.Fatal().Msgf("Program exiting: %s\n", err)
	}

	// if all the flags are provided via cli we ignore stdin by draining it
	if options.Stdin && (options.Domain != "" && options.ResolversFile != "" && options.Wordlist != "") {
		// drain stdin
		_, _ = io.Copy(io.Discard, os.Stdin)
		options.Stdin = false
	}

	// Set the domain in the config if provided by user from the stdin
	if options.Stdin && options.Wordlist != "" {
		buffer := &bytes.Buffer{}
		_, _ = io.Copy(buffer, os.Stdin)
		options.Domain = strings.TrimRight(buffer.String(), "\r\n")
	}

	return options
}

func createGroup(flagSet *goflags.FlagSet, groupName, description string, flags ...*goflags.FlagData) {
	flagSet.SetGroup(groupName, description)
	for _, currentFlag := range flags {
		currentFlag.Group(groupName)
	}
}
