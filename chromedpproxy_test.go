package chromedpproxy

import (
	"context"
	"errors"
	"github.com/chromedp/chromedp"
	"testing"
)

func TestClosingTwoTabsAndCreateOneTab(t *testing.T) {
	PrepareProxy(":9222", ":9221")
	target1ID, err := NewTab("about:blank")
	if err != nil {
		t.Error(err)
	}
	target2ID, err := NewTab("about:blank")
	if err != nil {
		t.Error(err)
	}
	err = CloseTarget(target1ID)
	if err != nil {
		t.Error(err)
	}
	err = CloseTarget(target2ID)
	if err != nil {
		t.Error(err)
	}

	PrepareProxy(":9222", ":9221")
	targetID, err := NewTab("about:blank")
	if err != nil {
		t.Error(err)
	}
	err = CloseTarget(targetID)
	if err != nil {
		t.Error(err)
	}
}
func TestDoubleClose(t *testing.T) {
	targetID, err := NewTab("about:blank")
	if err != nil {
		t.Error(err)
	}
	err = CloseTarget(targetID)
	if err != nil {
		t.Error(err)
	}
	err = CloseTarget(targetID)
	if err == nil {
		t.Error("expected error for attempting to close again not returned")
	}
}
func TestRegularChromeDPUsage(t *testing.T) {
	targetID, err := NewTab("https://github.com/jarylc/go-chromedpproxy")
	if err != nil {
		t.Error(err)
	}

	ctx := GetTarget(targetID)
	result := ""
	err = chromedp.Run(ctx, chromedp.Tasks{
		chromedp.InnerHTML(`strong[itemprop="name"] > a`, &result, chromedp.NodeVisible),
	})
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Error(err)
	}
	if result != "go-chromedpproxy" {
		t.Errorf("expected result to be 'go-chromedpproxy', got %s", result)
	}

	err = CloseTarget(targetID)
	if err != nil {
		t.Error(err)
	}
}
