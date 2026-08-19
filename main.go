package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/brandonapol/goodmorning/internal/config"
	"github.com/brandonapol/goodmorning/internal/source"
	"github.com/brandonapol/goodmorning/internal/tui"
)

func main() {
	configPath := flag.String("config", "", "path to config.json (default: OS config dir)")
	adhocPath := flag.String("path", "", "browse this single directory, ignoring config")
	flag.Parse()

	sources, err := buildSources(*configPath, *adhocPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "goodmorning:", err)
		os.Exit(1)
	}

	p := tea.NewProgram(tui.New(sources), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "goodmorning:", err)
		os.Exit(1)
	}
}

func buildSources(configPath, adhocPath string) ([]source.Source, error) {
	if adhocPath != "" {
		return []source.Source{source.NewFilesystemSource("Photos", adhocPath)}, nil
	}

	if configPath == "" {
		p, err := config.DefaultPath()
		if err != nil {
			return nil, fmt.Errorf("resolving config path: %w", err)
		}
		configPath = p
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}

	sources := make([]source.Source, len(cfg.Sources))
	for i, sc := range cfg.Sources {
		sources[i] = source.NewFilesystemSource(sc.Name, sc.Path)
	}
	return sources, nil
}
