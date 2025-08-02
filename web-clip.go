package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

var url string
var output string
var htmlFile string

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

func fullScreenshot(urlStr string, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		enableLifeCycleEvents(),
		navigateAndWaitFor(urlStr, "networkIdle"),
		chromedp.FullScreenshot(res, quality),
	}
}

func htmlScreenshot(htmlContent string, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		enableLifeCycleEvents(),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Load HTML content
			err := chromedp.Navigate("about:blank").Do(ctx)
			if err != nil {
				return err
			}
			
			// Set the HTML content
			err = chromedp.Evaluate(`document.documentElement.innerHTML = ` + "`" + htmlContent + "`", nil).Do(ctx)
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

func main() {
	flag.StringVar(&url, "url", "", "URL")
	flag.StringVar(&output, "output", "", "Output file path")
	flag.StringVar(&htmlFile, "html", "", "Local HTML file path")
	flag.Parse()

	if output == "" {
		fmt.Printf("Usage: %s [-url <URL> | -html <HTML_FILE>] -output <OUTPUT>\n", os.Args[0])
		return
	}

	if url == "" && htmlFile == "" {
		fmt.Printf("Error: Either -url or -html must be specified\n")
		fmt.Printf("Usage: %s [-url <URL> | -html <HTML_FILE>] -output <OUTPUT>\n", os.Args[0])
		return
	}

	if url != "" && htmlFile != "" {
		fmt.Printf("Error: Cannot specify both -url and -html\n")
		return
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var buf []byte
	var err error

	if url != "" {
		// Take screenshot from URL
		err = chromedp.Run(ctx, fullScreenshot(url, 90, &buf))
	} else {
		// Take screenshot from local HTML file
		htmlContent, err := readHTMLFile(htmlFile)
		if err != nil {
			panic(err)
		}
		err = chromedp.Run(ctx, htmlScreenshot(htmlContent, 90, &buf))
	}

	if err != nil {
		panic(err)
	}

	err = os.WriteFile(output, buf, 0o644)
	if err != nil {
		panic(err)
	}
}
