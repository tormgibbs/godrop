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

var readyConfirmationRegex = regexp.MustCompile(`INF Registered tunnel connection|INF Connected to Cloudflare|Tunnel is active at|tunnel is active at`)

func Start(ctx context.Context, port int, urlCh chan<- string) error {
	url := fmt.Sprintf("http://localhost:%d", port)
	cmd := exec.CommandContext(ctx, "cloudflared", "tunnel", "--url", url)

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
	var tunnelURL string
	wg.Add(2)

	scan := func(r io.Reader) {
		defer wg.Done()
		scanner := bufio.NewScanner(r)

		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		for scanner.Scan() {
			line := scanner.Text()

			if urlMatches := urlRegex.FindString(line); urlMatches != "" {
				tunnelURL = urlMatches
			}

			if tunnelURL != "" && readyConfirmationRegex.MatchString(line) {
				once.Do(func() {
					select {
					case urlCh <- tunnelURL:
					case <-ctx.Done():
					}
				})
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
