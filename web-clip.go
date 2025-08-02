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
		_, _, _, err := page.Navigate(url).Do(ctx)
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

func main() {
	flag.StringVar(&url, "url", "", "URL")
	flag.StringVar(&output, "output", "", "Output file path")
	flag.Parse()

	if url == "" || output == "" {
		fmt.Printf("Usage: %s -url <URL> -output <OUTPUT>\n", os.Args[0])
		return
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var buf []byte
	err := chromedp.Run(ctx, fullScreenshot(url, 90, &buf))
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(output, buf, 0o644)
	if err != nil {
		panic(err)
	}
}
