package main

import (
	"flag"
	"fmt"
	"github.com/rvk01/deflux/pkg/config"
	"github.com/rvk01/deflux/pkg/deflux"
	"log/slog"
	"os"
)

func main() {
	flagLoglevel := flag.String("loglevel", "warn", "debug | error | warn | info")
	flagConfigGen := flag.Bool("config-gen", false, "generate a default config and print it on stdout")
	flagConfig := flag.String("config", "", "specify the location of the config file (default: ./deflux.yml or /etc/deflux.yml)")
	flagOnce := flag.Bool("1", false, "write sensor state from REST API once and exit")
	flag.Parse()

	initLogging(flagLoglevel)

	if *flagConfigGen {
		config.OutputDefaultConfiguration()
		os.Exit(0)
	}

	cfg, err := config.LoadConfiguration(*flagConfig)
	if err != nil {
		slog.Error("No config file: %s", err)
		os.Exit(deflux.ExitFailConfig)
	}

	if *flagOnce {
		os.Exit(deflux.RunOnce(cfg))
	}

	os.Exit(deflux.RunWebsocket(cfg))
}

// initLogging initializes slog
func initLogging(flagLoglevel *string) {
	var logLevel = new(slog.LevelVar)
	err := logLevel.UnmarshalText([]byte(*flagLoglevel))
	if err != nil {
		fmt.Printf("Error parsing log level: %v", err)
		logLevel.Set(slog.LevelWarn)
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))
}
