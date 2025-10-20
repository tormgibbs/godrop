package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

func Start(
	ctx context.Context,
	file string,
	port int,
	ready chan<- struct{},
	downloadLimit int,
) error {
	mux := http.NewServeMux()
	filename := filepath.Base(file)

	var downloadCount int32

	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		if _, err := os.Stat(file); os.IsNotExist(err) {
			http.Error(w, "File not found or has been removed.", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Error accessing file.", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Disposition", "attachment; filename="+filename)
		w.Header().Set("Content-Type", "application/octet-stream")
		http.ServeFile(w, r, file)

		if downloadLimit > 0 {
			newCount := atomic.AddInt32(&downloadCount, 1)

			if newCount >= int32(downloadLimit) {
				cancel()
			}

		}

	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		close(ready)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		} else {
			errCh <- nil
		}
	}()

	select {
	case <-cancelCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil

	case err := <-errCh:
		return err
	}
}
