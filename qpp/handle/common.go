package handle

import (
	"auto-upload-product/config"
	"context"
	"fmt"

	"github.com/playwright-community/playwright-go"
)

func Login(ctx context.Context, page playwright.Page, config *config.Config) error {
	// 导航到登录页面
	if _, err := page.Goto(config.Login.URL); err != nil {
		return fmt.Errorf("导航到登录页面失败: %w", err)
	}

	// 等待登录元素加载完成
	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		return fmt.Errorf("等待登录元素加载失败: %w", err)
	}

	// 输入用户名和密码
	if err := page.Locator("input[name='account']").Fill(config.Login.Username); err != nil {
		return fmt.Errorf("输入用户名失败: %w", err)
	}
	if err := page.Locator("input[name='password']").Fill(config.Login.Password); err != nil {
		return fmt.Errorf("输入密码失败: %w", err)
	}

	return nil
}
