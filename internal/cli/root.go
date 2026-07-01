package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/taharaLovelace/Goverter/internal/buildinfo"
	"github.com/taharaLovelace/Goverter/internal/toolchain"
)

type dependencies struct {
	resolver toolchain.Resolver
}

func Execute(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	root := newRootCommand(dependencies{resolver: toolchain.NewResolver()})
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)

	err := root.ExecuteContext(ctx)
	if err == nil {
		return 0
	}
	if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
		fmt.Fprintln(stderr, "Conversion canceled.")
		return 130
	}
	var coded *exitError
	if errors.As(err, &coded) {
		fmt.Fprintf(stderr, "Error: %s\n", coded.err)
		return coded.code
	}
	fmt.Fprintf(stderr, "Error: %s\n", err)
	return 2
}

func newRootCommand(deps dependencies) *cobra.Command {
	root := &cobra.Command{
		Use:           "goverter",
		Short:         "Convert video, audio, and image files with FFmpeg",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       buildinfo.Version,
	}
	root.SetVersionTemplate(fmt.Sprintf(
		"goverter %s (commit %s, built %s)\n",
		buildinfo.Version,
		buildinfo.Commit,
		buildinfo.Date,
	))
	root.AddCommand(
		newConvertCommand(deps),
		newInfoCommand(deps),
		newFormatsCommand(),
	)
	root.InitDefaultCompletionCmd()
	return root
}
