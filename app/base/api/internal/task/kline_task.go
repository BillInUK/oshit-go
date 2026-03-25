package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/gofiber/fiber/v2/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"sync"
	"time"
)

type IntervalConfig struct {
	TimeUnit string // 时间单位（D/W/M）
	Quantity int    // 时间单位数量
}

// KLineTask 手续费统计任务
type KLineTask struct {
	db        *gorm.DB
	redis     redis.UniversalClient
	redSync   redsync.Redsync
	rpcClient *rpc.Client
	rpcURL    string
}

// KLineTask 创建手续费任务
func NewKLineTask(taskCtx *TaskContext) *KLineTask {
	rpcURL := taskCtx.ChainConfig.RPCURL
	rpcClient := rpc.New(rpcURL)
	return &KLineTask{
		db:        taskCtx.DB,
		redis:     taskCtx.Redis,
		redSync:   taskCtx.RedSync,
		rpcClient: rpcClient,
		rpcURL:    rpcURL,
	}
}

func (t *KLineTask) startKLineTask() {
	// 更新后的时间间隔配置
	intervals := map[string]IntervalConfig{
		"1D": {TimeUnit: "D", Quantity: 1},
		"1W": {TimeUnit: "W", Quantity: 1},
		"1M": {TimeUnit: "M", Quantity: 1},
	}

	for {
		// 并发获取所有间隔的数据
		var wg sync.WaitGroup
		for interval, config := range intervals {
			wg.Add(1)
			go func(intv string, cfg IntervalConfig) {
				defer wg.Done()
				t.getKlineWithBrowser(intv, cfg)
			}(interval, config)
		}
		wg.Wait()

		// 等待5小时后进入下一轮循环
		time.Sleep(5 * time.Hour)
	}
}

func (t *KLineTask) getKlineWithBrowser(interval string, config IntervalConfig) {
	var result string
	var timeFrom time.Time

	now := time.Now().UTC()
	timeTo := now.Unix()
	switch config.TimeUnit {
	case "D":
		timeFrom = now.AddDate(0, 0, -30*config.Quantity)
	case "W":
		timeFrom = now.AddDate(0, 0, -7*30*config.Quantity)
	case "M":
		timeFrom = now.AddDate(0, -30*config.Quantity, 0)
	default:
		log.Errorf("未知时间单位: %s", config.TimeUnit)
		return
	}

	mutexName := fmt.Sprintf("FETCH-BIRD-EYE-APP-PRICE-%s", interval)
	scanMutex := t.redSync.NewMutex(mutexName)

	if err := scanMutex.Lock(); err != nil {
		var errTaken *redsync.ErrTaken
		if errors.As(err, &errTaken) {
			return
		}
		log.Errorf("获取%s锁失败: %v", interval, err)
		return
	}
	defer scanMutex.Unlock()

	bc, err := t.browserInitRaydium()
	if err != nil {
		log.Errorf("获取Bird Eye Data错误，创建浏览器上下文错误: %v", err)
		return
	}
	defer bc.Close()
	evalCmd := fmt.Sprintf(`
			fetch('https://birdeye-proxy.raydium.io/defi/ohlcv/base_quote?base_address=ShitJuMfPKCQU7LedLERFYapDta7CCdKExPWX2gETRH&quote_address=So11111111111111111111111111111111111111112&type=1D&time_from=%s&time_to=%s', {
				headers: {
					'authority': 'birdeye-proxy.raydium.io',
					'method': 'GET',
					'path': '/defi/ohlcv/base_quote?base_address=ShitJuMfPKCQU7LedLERFYapDta7CCdKExPWX2gETRH&quote_address=So11111111111111111111111111111111111111112&type=1D&time_from=1724976000&time_to=1750982400',
					'scheme': 'https',
					'accept': 'application/json, text/plain, */*',
					'accept-encoding': 'gzip, deflate, br, zstd',
					'accept-language': 'zh-CN,zh;q=0.9,en;q=0.8',
					'cache-control': 'no-cache',
					'origin': 'https://raydium.io',
					'pragma': 'no-cache',
					'priority': 'u=1, i',
					'referer': 'https://raydium.io/',
					'sec-ch-ua': '"Google Chrome";v="137", "Chromium";v="137", "Not/A)Brand";v="24"',
					'sec-ch-ua-mobile': '?0',
					'sec-ch-ua-platform': '"macOS"',
					'sec-fetch-dest': 'empty',
					'sec-fetch-mode': 'cors',
					'sec-fetch-site': 'same-site',
					'user-agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36'
				},
				credentials: 'include'
			})
			.then(response => {
				if (!response.ok) {
					throw new Error('HTTP error! status: ' + response.status);
				}
				return response.text();
			})
			.then(data => {
				window.__klineResult = data;
				console.log('K线数据获取成功:', data);
			})
			.catch(error => {
				console.error('K线数据获取失败:', error);
				window.__klineResult = 'Error: ' + error.message;
			})
		`, strconv.FormatInt(timeFrom.Unix(), 10),
		strconv.FormatInt(timeTo, 10))

	err = chromedp.Run(bc.Ctx,
		chromedp.EvaluateAsDevTools(evalCmd, nil),
		chromedp.Sleep(5*time.Second), // 等待请求完成
		chromedp.EvaluateAsDevTools(`window.__klineResult || "No data"`, &result),
	)

	log.Trace("获取BirdEye原始数据: %v", result)

	// 解析并格式化JSON
	var rawData interface{}
	if err := json.Unmarshal([]byte(result), &rawData); err != nil {
		log.Errorf("解析%s JSON失败: %v", interval, err)
		return
	}

	formattedData, err := json.MarshalIndent(rawData, "", "  ")
	if err != nil {
		log.Errorf("格式化%s JSON失败: %v", interval, err)
		return
	}

	// 存储到Redis
	redisKey := fmt.Sprintf("BIRD-EYE-APP-PRICE-%s", interval)
	if err := t.redis.Set(
		context.Background(),
		redisKey,
		string(formattedData),
		12*time.Hour,
	).Err(); err != nil {
		log.Errorf("%s Redis存储失败: %v", interval, err)
		return
	}

	log.Infof("%s数据更新成功", interval)

	return
}

// 获取浏览器上下文中的所有 Cookie
func (t *KLineTask) getCookies(ctx context.Context) (string, error) {
	// 获取当前域名下的所有 cookies
	cookies, err := network.GetCookies().Do(ctx)
	if err != nil {
		return "", fmt.Errorf("获取 cookies 失败: %v", err)
	}

	var cookieStrings []string
	for _, cookie := range cookies {
		cookieStrings = append(cookieStrings, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}

	return strings.Join(cookieStrings, "; "), nil
}

func (t *KLineTask) browserInitRaydium() (*BrowserContext, error) {
	// 创建上下文
	bc, _ := CreateBrowserContext(true)
	ctx := bc.Ctx

	// 目标URL
	url := "https://raydium.io/swap/?inputMint=sol&outputMint=ShitJuMfPKCQU7LedLERFYapDta7CCdKExPWX2gETRH"

	// 用于存储 cookies
	var cookies string

	err := chromedp.Run(ctx,
		// 导航到目标URL
		chromedp.Navigate(url),

		// 等待页面加载完成
		chromedp.WaitReady("body"),

		// ========== 处理第一个弹窗 (ConfirmToken) ==========
		// 等待第一个弹窗的按钮出现 - 使用文本内容匹配
		chromedp.WaitVisible(`//button[contains(text(), "I understand, confirm")]`, chromedp.BySearch),

		// 点击第一个弹窗的确认按钮
		chromedp.Click(`//button[contains(text(), "I understand, confirm")]`, chromedp.BySearch),

		// 等待1秒，让页面稳定
		chromedp.Sleep(1*time.Second),

		// ========== 处理 Invalid rpc node 弹窗 ==========
		chromedp.ActionFunc(func(ctx context.Context) error {
			// 检查是否出现 Invalid rpc node 弹窗
			var nodes []*cdp.Node
			if err := chromedp.Nodes(`//div[contains(text(), "Invalid rpc node")]`, &nodes, chromedp.AtLeast(0), chromedp.BySearch).Do(ctx); err != nil {
				return err
			}

			if len(nodes) > 0 {
				log.Tracef("检测到 Invalid rpc node 弹窗，等待10秒让其自动消失...")
				// 等待10秒让弹窗自动消失
				chromedp.Sleep(10 * time.Second).Do(ctx)
				log.Tracef("已等待10秒，继续执行后续流程")
			}
			return nil
		}),

		// ========== 处理第二个弹窗 (Disclaimer) ==========
		// 等待第二个弹窗出现
		chromedp.WaitVisible(`//p[contains(text(), "Agree to terms")]`, chromedp.BySearch),

		// 定位并点击复选框
		chromedp.ActionFunc(func(ctx context.Context) error {
			// 找到包含 "Agree to terms" 文本的 <p> 元素
			var pNodes []*cdp.Node
			if err := chromedp.Nodes(`//p[contains(text(), "Agree to terms")]`, &pNodes, chromedp.AtLeast(1), chromedp.BySearch).Do(ctx); err != nil {
				return err
			}

			if len(pNodes) == 0 {
				return fmt.Errorf("未找到 'Agree to terms' 文本")
			}

			// 找到最近的复选框
			var checkboxNodes []*cdp.Node
			if err := chromedp.Nodes(`//span[contains(@class, "chakra-checkbox__control")]`, &checkboxNodes, chromedp.AtLeast(1), chromedp.BySearch).Do(ctx); err != nil {
				return err
			}

			if len(checkboxNodes) == 0 {
				return fmt.Errorf("未找到复选框")
			}

			log.Tracef("点击复选框...")
			return chromedp.MouseClickNode(checkboxNodes[0]).Do(ctx)
		}),

		// 等待进入按钮变为可用状态
		chromedp.ActionFunc(func(ctx context.Context) error {
			timeout := time.After(20 * time.Second)
			ticker := time.NewTicker(500 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-timeout:
					return fmt.Errorf("等待按钮可用超时")
				case <-ticker.C:
					// 检查按钮是否可用
					var isDisabled bool
					if err := chromedp.EvaluateAsDevTools(`document.querySelector('button[data-sentry-source-file="DisclaimerModal.tsx"]').disabled`, &isDisabled).Do(ctx); err != nil {
						return err
					}

					if !isDisabled {
						return nil
					}
				}
			}
		}),

		// 点击进入按钮
		chromedp.Click(`//button[contains(text(), "Enter Raydium")]`, chromedp.BySearch),

		// 等待主页面加载完成
		chromedp.WaitVisible(`//input[@name="swap" and @data-sentry-source-file="TokenInput.tsx"]`, chromedp.BySearch),

		// 获取cookies并保存到变量
		chromedp.ActionFunc(func(ctx context.Context) error {
			c, err := t.getCookies(ctx)
			if err != nil {
				return err
			}
			cookies = c
			log.Tracef("成功获取Cookies: %s", cookies)
			return nil
		}),
	)
	return bc, err
}
