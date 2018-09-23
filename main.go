package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	outputFormat     string
	initialAddresses string
	password         string
)

var outputs = map[string]Outputter{
	"debug":    newDebugOutputter(os.Stdout),
	"graphviz": newGraphvizOutputter(os.Stdout),
}

func init() {
	flag.StringVar(&outputFormat, "output", "graphviz", "the output format")
	flag.StringVar(&initialAddresses, "addr", "", "the initial addresses, csv style")
	flag.StringVar(&password, "pass", "", "the authentication password")
}

func main() {
	flag.Parse()

	addresses := strings.Split(initialAddresses, ",")
	if len(addresses) == 0 {
		fmt.Println("Flag addr not set.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	discoverer := NewDiscoverer(addresses, password)

	if err := discoverer.BuildGraph(); err != nil {
		fmt.Printf("During building of graph %v", err)
		os.Exit(1)
	}

	if outputter, ok := outputs[outputFormat]; ok {
		outputter.Print(discoverer.Result())
	} else {
		fmt.Println("Unknown output")
		os.Exit(1)
	}
}
