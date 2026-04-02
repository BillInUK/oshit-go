package task

import (
	"context"
	"github.com/chromedp/chromedp"
	"time"
)

type BrowserContext struct {
	AllocCtx      context.Context
	AllocCancel   context.CancelFunc
	Ctx           context.Context
	chromedpCancel context.CancelFunc // chromedp.NewContext 的 cancel，确保浏览器进程正常退出
	timeoutCancel  context.CancelFunc // context.WithTimeout 的 cancel，释放 timer 资源
}

// CreateBrowserContext 创建浏览器上下文资源
func CreateBrowserContext(headless bool) (*BrowserContext, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", headless), // 调试时可设为false
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	chromedpCtx, chromedpCancel := chromedp.NewContext(allocCtx)

	// 在 chromedp context 上叠加超时，两个 cancel 分开保存
	ctx, timeoutCancel := context.WithTimeout(chromedpCtx, 1*time.Minute)

	return &BrowserContext{
		AllocCtx:      allocCtx,
		AllocCancel:   allocCancel,
		Ctx:           ctx,
		chromedpCancel: chromedpCancel,
		timeoutCancel:  timeoutCancel,
	}, nil
}

func (bc BrowserContext) Close() {
	bc.timeoutCancel()   // 释放 timer
	bc.chromedpCancel()  // 关闭 chromedp context，触发浏览器进程退出
	bc.AllocCancel()     // 释放 allocator，清理临时 profile 目录
}
