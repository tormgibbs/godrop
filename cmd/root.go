package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
	"github.com/tormgibbs/godrop/internal/server"
	"github.com/tormgibbs/godrop/internal/tunnel"
	"github.com/tormgibbs/godrop/internal/util"
	"golang.org/x/sync/errgroup"
)

var rootCmd = &cobra.Command{
	Use:   "godrop [file|directory]",
	Short: "Share a file securely over a temporary Cloudflare tunnel",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("missing path argument")
		}

		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			return err
		}

		limit, err := cmd.Flags().GetInt("limit")
		if err != nil {
			return err
		}

		if limit == 0 {
			once, err := cmd.Flags().GetBool("once")
			if err != nil {
				return err
			}
			if once {
				limit = 1
			}
		}

		path, err := util.ExpandPath(args[0])
		if err != nil {
			return fmt.Errorf("failed to expand path: %w", err)
		}

		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("path %q doesn’t exist", path)
			}
			return fmt.Errorf("error checking path: %w", err)
		}

		var fileToRemove string

		if info.IsDir() {
			zipFileName := filepath.Join(".", filepath.Base(path)+".zip")
			if err := util.ZipDirectory(path, zipFileName); err != nil {
				return fmt.Errorf("failed to zip directory: %w", err)
			}
			path = zipFileName

			fileToRemove = zipFileName
			defer func() {
				if fileToRemove != "" {
					if err := os.Remove(fileToRemove); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: Failed to remove temporary file %q: %v\n", fileToRemove, err)
					}
				}
			}()
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		group, ctx := errgroup.WithContext(ctx)

		ready := make(chan struct{})

		urlCh := make(chan string, 1)

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Preparing your file and establishing a secure tunnel..."
		s.Start()

		group.Go(func() error {
			return server.Start(ctx, path, port, ready, limit)
		})

		group.Go(func() error {
			<-ready
			return tunnel.Start(ctx, port, urlCh)
		})

		go func() {
			select {
			case url := <-urlCh:
				s.Stop()
				fmt.Printf("Your file is ready at: %s\n\n", url)
				fmt.Println("This link is temporary and secure. Press Ctrl+C to stop sharing")
			case <-ctx.Done():
				s.Stop()
			}
		}()

		if err := group.Wait(); err != nil {
			return fmt.Errorf("run failed: %w", err)
		}

		fmt.Println("All services stopped. Your file is no longer being shared")

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("once", "o", false, "serve once and exit")
	rootCmd.Flags().IntP("port", "p", 8080, "port to listen on")
	rootCmd.Flags().IntP("limit", "l", 0, "maximum number of downloads before shutting down (0 means no limit)")
	rootCmd.SilenceUsage = true
}
