// Command hopscope is a local network diagnostics agent that explains why an
// internet connection feels slow, layer by layer.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dasosis/hopscope/internal/cli"
)

func main() {
	os.Exit(run())
}

// run holds the real body of main so that deferred calls execute before the
// process exits; os.Exit would otherwise skip them.
func run() int {
	// Cancel the root context on Ctrl-C or SIGTERM so long-running probes can
	// stop cleanly and still print partial results.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := cli.NewRootCmd().ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "hopscope:", err)
		return 1
	}
	return 0
}
