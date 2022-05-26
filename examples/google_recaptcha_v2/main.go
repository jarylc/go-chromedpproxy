package main

import (
	"context"
	"errors"
	"github.com/chromedp/chromedp"
	"github.com/jarylc/go-chromedpproxy"
	"log"
)

func main() {
	// create result channel to signify end of program
	resultChan := make(chan string)

	// prepare proxy using port 9222 as Chrome debugging port and 9221 as frontend port
	// also send in `chromedp.DisableGPU` as extra ExecAllocatorOption
	chromedpproxy.PrepareProxy(":9222", ":9221", chromedp.DisableGPU)
	targetID, err := chromedpproxy.NewTab("https://www.google.com/recaptcha/api2/demo", chromedp.WithLogf(log.Printf))
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Panic(err)
	}
	defer chromedpproxy.CloseTarget(targetID)

	ctx := chromedpproxy.GetTarget(targetID)

	// parallel steps to check if a captcha really exists before notifying
	go func() {
		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Click(`.recaptcha-checkbox-border`, chromedp.NodeVisible),
			chromedp.WaitVisible(`#rc-imageselect`, chromedp.NodeVisible),
			chromedp.ActionFunc(func(ctx context.Context) error {
				log.Print("You would normally send this via any form of notifications to a user:")
				log.Printf("Recaptcha detected, please solve it here: http://127.0.0.1:9221/?id=%s", targetID)
				return nil
			}),
		})
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Panic(err)
		}
	}()

	// parallel steps to auto submit once captcha is solved
	go func() {
		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.WaitVisible(`.recaptcha-checkbox-checked`),
			chromedp.Click(`#recaptcha-demo-submit`),
		})
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Panic(err)
		}
	}()

	// parallel steps to find success condition, once form is submitted regardless of captcha
	go func() {
		var result = "Failed"
		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.InnerHTML(`.recaptcha-success`, &result, chromedp.NodeVisible),
			chromedp.ActionFunc(func(ctx context.Context) error {
				log.Print("Captcha solved!")
				return nil
			}),
		})
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Panic(err)
		}
		resultChan <- result
	}()

	// you can do other stuff here like grab cookie data, etc.

	// log and close
	log.Print("Result: ", <-resultChan)
}
