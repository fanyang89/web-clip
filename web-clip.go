package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

var output string
var dpi int
var input string

func enableLifeCycleEvents() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		err := page.Enable().Do(ctx)
		if err != nil {
			return err
		}
		err = page.SetLifecycleEventsEnabled(true).Do(ctx)
		if err != nil {
			return err
		}
		return nil
	}
}

func setDeviceScaleFactor(scaleFactor float64) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		err := emulation.SetDeviceMetricsOverride(0, 0, scaleFactor, false).Do(ctx)
		return err
	}
}

func waitFor(ctx context.Context, eventName string) error {
	ch := make(chan struct{})
	cctx, cancel := context.WithCancel(ctx)
	chromedp.ListenTarget(cctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *page.EventLifecycleEvent:
			if e.Name == eventName {
				cancel()
				close(ch)
			}
		}
	})
	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func navigateAndWaitFor(url string, eventName string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		_, _, _, _, err := page.Navigate(url).Do(ctx)
		if err != nil {
			return err
		}

		return waitFor(ctx, eventName)
	}
}

func fullScreenshot(urlStr string, quality int, dpi int, res *[]byte) chromedp.Tasks {
	scaleFactor := float64(dpi) / 96.0 // Convert DPI to scale factor (96 DPI = 1.0)
	return chromedp.Tasks{
		enableLifeCycleEvents(),
		setDeviceScaleFactor(scaleFactor),
		navigateAndWaitFor(urlStr, "networkIdle"),
		chromedp.FullScreenshot(res, quality),
	}
}

func htmlScreenshot(htmlContent string, quality int, dpi int, res *[]byte) chromedp.Tasks {
	scaleFactor := float64(dpi) / 96.0 // Convert DPI to scale factor (96 DPI = 1.0)
	return chromedp.Tasks{
		enableLifeCycleEvents(),
		setDeviceScaleFactor(scaleFactor),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Load HTML content
			err := chromedp.Navigate("about:blank").Do(ctx)
			if err != nil {
				return err
			}

			// Set the HTML content
			err = chromedp.Evaluate(`document.documentElement.innerHTML = `+"`"+htmlContent+"`", nil).Do(ctx)
			if err != nil {
				return err
			}

			// Wait for the page to be ready
			return waitFor(ctx, "networkIdle")
		}),
		chromedp.FullScreenshot(res, quality),
	}
}

func readHTMLFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func isURL(input string) bool {
	return strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://")
}

func isFile(input string) bool {
	_, err := os.Stat(input)
	return err == nil
}

func main() {
	flag.StringVar(&output, "output", "", "Output file path")
	flag.IntVar(&dpi, "dpi", 200, "DPI for screenshot (default: 200)")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Printf("Error: Exactly one input argument (URL or HTML file path) is required\n")
		fmt.Printf("Usage: %s <URL_OR_HTML_FILE> [-output <OUTPUT>]\n", os.Args[0])
		return
	}

	input = flag.Arg(0)

	if !isURL(input) && !isFile(input) {
		fmt.Printf("Error: Input must be a valid URL (http:// or https://) or an existing HTML file path\n")
		return
	}

	if output == "" {
		if isFile(input) {
			dir := filepath.Dir(input)
			baseName := strings.TrimSuffix(filepath.Base(input), filepath.Ext(input))
			output = filepath.Join(dir, baseName+".png")
		} else {
			output = "screenshot.png"
		}
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var buf []byte
	var err error

	if isURL(input) {
		// Take a screenshot from URL
		err = chromedp.Run(ctx, fullScreenshot(input, 90, dpi, &buf))
	} else {
		// Take a screenshot from local HTML file
		htmlContent, err := readHTMLFile(input)
		if err != nil {
			panic(err)
		}
		err = chromedp.Run(ctx, htmlScreenshot(htmlContent, 90, dpi, &buf))
	}

	if err != nil {
		panic(err)
	}

	err = os.WriteFile(output, buf, 0o644)
	if err != nil {
		panic(err)
	}
}
