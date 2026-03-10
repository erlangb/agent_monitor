// Command inspector is a standalone TUI for browsing agent use cases and inspecting runs.
// Usage: go run ./cmd/inspector
package main

import (
	"context"
	"erlangb/agentmonitor/internal/application"
	"erlangb/agentmonitor/internal/application/terminal"
	"erlangb/agentmonitor/internal/application/tui"
	"erlangb/agentmonitor/internal/config"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = filepath.Join("config", "config.yaml")
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		slog.Default().Error("config load failed", "error", err)
		os.Exit(1)
	}

	runnerFlag := flag.String("runner", "tea", "runner to use: tea | terminal")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var runner application.Runner
	switch *runnerFlag {
	case "terminal":
		runner = terminal.New()
	default:
		runner = tui.NewAgentMonitorTeaRunner()
	}

	app, err := application.NewApplication(ctx, *cfg, runner)
	if err != nil {
		slog.Default().Error("application init failed", "error", err)
		return
	}
	defer app.Close()

	err = app.Run(ctx)
	if err != nil {
		slog.Default().Error("application error", "error", err)
		return
	}

	slog.Default().Info("application exited")
}
