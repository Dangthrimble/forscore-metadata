package namesedit

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"fn2fm/internal/names"
)

// Run starts the local dictionary editor. args may contain an optional path.
func Run(args []string) error {
	var flagPath string
	if len(args) > 1 {
		return fmt.Errorf("usage: fn2fm names [path]")
	}
	if len(args) == 1 {
		flagPath = args[0]
	}
	path, err := names.ResolvePath(flagPath)
	if err != nil {
		return err
	}
	cfg, err := names.LoadConfig()
	if err != nil {
		return err
	}
	fromConfig := strings.TrimSpace(flagPath) == "" && cfg.NamesPath != ""

	var dict *names.Dictionary
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if fromConfig {
			return fmt.Errorf("configured dictionary not found: %s", path)
		}
		dict = names.NewEmpty(path)
	} else {
		dict, err = names.Load(path)
		if err != nil {
			return err
		}
	}

	token, err := randomToken()
	if err != nil {
		return err
	}
	srv := &server{dict: dict, token: token}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://%s/?token=%s", ln.Addr().String(), token)

	httpServer := &http.Server{Handler: srv.routes(), ReadHeaderTimeout: 5 * time.Second}
	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.Serve(ln)
	}()

	fmt.Printf("Dictionary editor for %s\n", dict.Path)
	fmt.Printf("Open %s\n", url)
	fmt.Println("Press Ctrl+C to stop.")
	_ = openBrowser(url)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	select {
	case <-sig:
		fmt.Println("\nStopping editor.")
		_ = httpServer.Close()
		return nil
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
