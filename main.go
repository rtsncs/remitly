package main

import (
	"flag"
	"log"
	"os"

	"github.com/rtsncs/remitly-swift-api/loader"
	"github.com/rtsncs/remitly-swift-api/server"
)

func main() {
	loadCmd := flag.NewFlagSet("load", flag.ExitOnError)
	loadFile := loadCmd.String("file", "", "Path to the SWIFT data spreadsheet")

	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)

	if len(os.Args) < 2 {
		log.Fatalln("expected 'load' or 'serve' subcommand")
	}

	switch os.Args[1] {
	case "load":
		loadCmd.Parse(os.Args[2:])
		if *loadFile == "" {
			log.Fatalf("Usage: %s load -file=path/to/file.xlsx\n", os.Args[0])
		}
		if err := loader.LoadFromFile(*loadFile); err != nil {
			log.Fatal(err)
		}
	case "serve":
		serveCmd.Parse(os.Args[2:])
		server.Run()
	default:
		log.Fatalln("expected 'load' or 'serve' subcommand")
	}
}
