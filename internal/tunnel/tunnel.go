package tunnel

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"sync"
)

var urlRegex = regexp.MustCompile(`https://[a-zA-Z0-9\-]+\.trycloudflare\.com`)

func Start(ctx context.Context, urlCh chan<- string) error {
	cmd := exec.CommandContext(ctx, "cloudflared", "tunnel", "--url", "http://localhost:8080")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start cloudflared: %w", err)
	}

	var wg sync.WaitGroup
	var once sync.Once
	wg.Add(2)

	scan := func(r io.Reader) {
		defer wg.Done()
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			if matches := urlRegex.FindString(line); matches != "" {
				once.Do(func() {
					select {
					case urlCh <- matches:
					case <-ctx.Done():
					}
				})
				return
			}
		}
	}

	go scan(stdout)
	go scan(stderr)

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			wg.Wait()
			return nil
		}

		wg.Wait()
		return fmt.Errorf("cloudflared exited: %w", err)
	}

	wg.Wait()
	return nil
}
