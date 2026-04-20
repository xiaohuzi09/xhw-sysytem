package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"auto-upload-product/config"
	"auto-upload-product/handle"

	"github.com/playwright-community/playwright-go"
)

func main() {
	// 加载配置文件
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Printf("加载配置文件失败，使用默认配置: %v", err)
		cfg = config.GetDefaultConfig()
	}

	// 打印配置信息
	fmt.Printf("浏览器配置: Headless=%v, Timeout=%d秒\n", cfg.Browser.Headless, cfg.Browser.Timeout)
	fmt.Printf("登录配置: URL=%s, WaitTime=%d秒\n", cfg.Login.URL, cfg.Login.WaitTime)
	fmt.Printf("存储配置: StatePath=%s, ScreenshotPath=%s\n", cfg.Storage.StatePath, cfg.Storage.ScreenshotPath)
	fmt.Printf("自动化配置: RetryCount=%d, RetryDelay=%d秒\n", cfg.Automation.RetryCount, cfg.Automation.RetryDelay)

	// 安装Playwright浏览器
	if err := playwright.Install(); err != nil {
		log.Fatalf("安装Playwright浏览器失败: %v", err)
	}

	ctx := context.Background()

	// 启动浏览器
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("启动playwright失败: %v", err)
	}
	// defer pw.Stop()

	// 创建浏览器上下文
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(cfg.Browser.Headless),
	})
	if err != nil {
		log.Fatalf("启动浏览器失败: %v", err)
	}
	// defer browser.Close()

	// 创建新页面
	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("创建新页面失败: %v", err)
	}
	// defer page.Close()

	// 执行登录（填写用户名密码）
	if err := handle.Login(ctx, page, cfg); err != nil {
		log.Fatalf("登录失败: %v", err)
	}

	// 等待用户手动点击登录按钮
	fmt.Println("\n请在浏览器中手动点击登录按钮...")
	fmt.Println("登录成功后按回车键继续执行...")
	bufio.NewReader(os.Stdin).ReadString('\n')

	// 继续执行：导航到产品页面
	fmt.Println("正在导航到产品页面...")
	if err := handle.ToProductPage(ctx, page, cfg); err != nil {
		log.Fatalf("导航到产品页面失败: %v", err)
	}

	// 等待页面加载完成
	fmt.Println("等待页面加载...")
	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		log.Printf("等待页面加载警告: %v", err)
	}
	fmt.Println("已到达产品页面！")

	product := handle.Product{TitleCh: "测试产品", TitleEn: "Test Product", MaterialImg: "./image/022631.png"}

	// 填写产品信息
	if err := handle.FillProductInfo(ctx, page, product); err != nil {
		log.Printf("填写产品信息失败: %v", err)
	} else {
		fmt.Println("产品信息填写成功！")
	}

	// 等待用户按回车键退出，保持浏览器窗口打开
	fmt.Println("\n按回车键退出...")
	bufio.NewReader(os.Stdin).ReadString('\n')
}
