package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"passwordmanager/cmd/client/api"
	"passwordmanager/cmd/client/ui"

	tea "charm.land/bubbletea/v2"
)

var (
	version   = "dev"
	buildDate = "unknown"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	serverURL := flag.String("server", "https://localhost:443", "server URL")
	flag.StringVar(serverURL, "s", "https://localhost:443", "server URL shorthand")
	showVersion := flag.Bool("version", false, "print client version and build date")
	flag.BoolVar(showVersion, "v", false, "print client version and build date (shorthand)")
	flag.Parse()

	if *showVersion {
		fmt.Printf("passwordmanager-client %s (built %s)\n", version, buildDate)
		return
	}

	client := api.NewClient(*serverURL)
	p := tea.NewProgram(
		ui.InitialModel(client),
	)
	if _, err := p.Run(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
