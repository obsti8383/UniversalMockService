package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const configFile = "config.json"
const Timeout = 5 * time.Second // Timeout is the amount of time the server will wait for requests to finish during shutdown

// Configuration is the struct that gets filled by reading config.json JSON file
type Configuration struct {
	VerboseOutput       bool   `json:"verbose"`
	InterfaceAndPort    string `json:"interfaceAndPort"`
	ResponseFile        string `json:"responseFile"`
	ResponseContentType string `json:"responseContentType"`
}

func initConfig() (configuration Configuration, err error) {
	// get configuration from config json
	file, err := os.Open(configFile)
	if err != nil {
		return configuration, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configuration)
	if err != nil {
		return configuration, err
	}

	return configuration, nil
}

func main() {
	errorLogger := log.New(os.Stderr, "", 0)
	debugLogger := log.New(io.Discard, "", 0)

	configuration, err := initConfig()
	if err != nil {
		// remove Fatal() in case config JSON file is optional
		debugLogger.Println(err.Error())
	}
	// check for mandatory configuration variables
	if configuration.InterfaceAndPort == "" {
		configuration.InterfaceAndPort = ":50000"
	}
	if configuration.ResponseFile == "" {
		configuration.ResponseFile = "response.txt"
	}
	if configuration.ResponseContentType == "" {
		configuration.ResponseContentType = "text/xml; charset=UTF-8"
	}

	// evaluate command line flags
	var help bool
	var verbose bool
	flags := flag.NewFlagSet("UniversalMockService", flag.ContinueOnError)
	flags.BoolVar(&help, "help", false, "Show this help message")
	flags.BoolVar(&help, "h", false, "")
	flags.BoolVar(&verbose, "v", configuration.VerboseOutput, "Show verbose logging.")
	flags.StringVar(&configuration.InterfaceAndPort, "interfaceAndPort", configuration.InterfaceAndPort, "interface and port e.g. localhost:50000 or :50000 for all interfaces")
	flags.StringVar(&configuration.ResponseFile, "responseFile", configuration.ResponseFile, "the file that will be sent as response to every request")
	flags.StringVar(&configuration.ResponseContentType, "responseContentType", configuration.ResponseContentType, "the Content-Type response header")
	err = flags.Parse(os.Args[1:])
	switch err {
	case flag.ErrHelp:
		help = true
	case nil:
	default:
		errorLogger.Fatalf("error parsing flags: %v", err)
	}
	// If the help flag was set, just show the help message and exit.
	if help {
		printHelp(flags)
		os.Exit(0)
	}

	// check if response file exists before starting server
	if !fileExists(configuration.ResponseFile) {
		errorLogger.Println("Response file", configuration.ResponseFile, "does not exist or is a directory")
		os.Exit(1)
	}

	// init server struct
	srv := &http.Server{Addr: configuration.InterfaceAndPort, Handler: &App{configuration, errorLogger, debugLogger}}

	// subscribe to SIGINT signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// start listener
	errc := make(chan error)
	go func() {
		errorLogger.Println("Starting mock service with interface \"" + configuration.InterfaceAndPort + "\" response file \"" + configuration.ResponseFile + "\" and response Content-Type \"" + configuration.ResponseContentType + "\"")
		// service connections
		errc <- srv.ListenAndServe()
	}()

	<-stopChan // wait for system signal
	errorLogger.Println("Shutting down server...")

	// shut down gracefully, but wait no longer than 5 seconds before halting
	ctx, c := context.WithTimeout(context.Background(), Timeout)
	defer c()
	srv.Shutdown(ctx)

	select {
	case err := <-errc:
		errorLogger.Printf("Finished listening: %v\n", err)
	case <-ctx.Done():
		errorLogger.Println("Graceful shutdown timed out")
	}

	errorLogger.Println("Server stopped")
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// printHelp prints command line parameter help
func printHelp(flags *flag.FlagSet) {
	fmt.Fprintf(flags.Output(), "\nUsage of %s:\n", os.Args[0])
	flags.PrintDefaults()
	fmt.Printf(`

To configure UniversalMockService you can also use a config.json file. Example:

	{
		"verbose": false,
		"interfaceAndPort": "localhost:20000",
		"responseFile": "response2.txt",
		"responseContentType": "text/xml; charset=UTF-8"
	}
`)
}
