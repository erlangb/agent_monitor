package terminal

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"erlangb/agentmonitor/internal/usecase"

	"github.com/erlangb/agentmeter"
	"github.com/erlangb/agentmeter/reasoning"
)

// TerminalRunner implements application.Runner as a simple stdin/stdout REPL.
type TerminalRunner struct {
	lines chan string
}

// New returns a TerminalRunner ready to drive a stdin/stdout REPL.
func New() *TerminalRunner { return &TerminalRunner{lines: make(chan string)} }

func (r *TerminalRunner) Run(ctx context.Context, useCases []usecase.UseCase, meter *agentmeter.Meter) error {
	// Read stdin in a background goroutine so we can select on ctx.Done().
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			r.lines <- scanner.Text()
		}
		close(r.lines)
	}()

	for {
		uc, err := r.selectUseCase(ctx, useCases)
		if err != nil {
			return err
		}
		if uc == nil {
			fmt.Println("\nbye.")
			return nil
		}
		if err := r.runLoop(ctx, uc, meter); err != nil {
			return err
		}
	}
}

func (r *TerminalRunner) Close() {}

// readLine blocks until a line is available or ctx is cancelled.
func (r *TerminalRunner) readLine(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case line, ok := <-r.lines:
		if !ok {
			return "", fmt.Errorf("stdin closed")
		}
		return line, nil
	}
}

// selectUseCase prints the menu and returns the chosen use case, or nil to quit.
func (r *TerminalRunner) selectUseCase(ctx context.Context, useCases []usecase.UseCase) (usecase.UseCase, error) {
	fmt.Println("\n--- use cases ---")
	for i, uc := range useCases {
		fmt.Printf("  %d. %s — %s\n", i+1, uc.Name(), uc.Description())
	}
	fmt.Print("  0. quit\n> ")

	line, err := r.readLine(ctx)
	if err != nil {
		return nil, err
	}
	choice, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || choice < 0 || choice > len(useCases) {
		fmt.Println("invalid choice")
		return r.selectUseCase(ctx, useCases)
	}
	if choice == 0 {
		return nil, nil
	}
	return useCases[choice-1], nil
}

// runLoop runs a query loop for the selected use case until the user types "back".
func (r *TerminalRunner) runLoop(ctx context.Context, uc usecase.UseCase, meter *agentmeter.Meter) error {
	fmt.Printf("\n[%s] type 'back' to return to menu\n", uc.Name())
	if ex := uc.ExampleInput(); ex != "" {
		fmt.Printf("example: %s\n", ex)
	}

	printer := reasoning.NewPrinter(os.Stdout)

	for {
		fmt.Print("> ")
		line, err := r.readLine(ctx)
		if err != nil {
			return err
		}
		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}
		if strings.EqualFold(input, "back") {
			meter.ClearHistory()
			return nil
		}

		result, err := uc.Run(ctx, input)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			continue
		}

		fmt.Println(result)
		printer.PrintHistory(meter.History())
	}
}
