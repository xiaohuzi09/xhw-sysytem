package handle

import (
	"auto-upload-product/config"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/playwright-community/playwright-go"
)

type Product struct {
	TitleCh     string   `json:"titleCh"`
	TitleEn     string   `json:"titleEn"`
	ImagePaths  []string `json:"imagePaths"` // 本地图片路径列表
	MaterialImg string   `json:"materialImg"`
}

// 导航到产品页面
func ToProductPage(ctx context.Context, page playwright.Page, config *config.Config) error {
	if _, err := page.Goto("https://www.dianxiaomi.com/web/popTemu/quoteEdit?id=154449843051332182"); err != nil {
		return fmt.Errorf("导航到产品页面失败: %w", err)
	}
	return nil
}

func FillProductInfo(ctx context.Context, page playwright.Page, product Product) error {
	// 填写中文标题（简化选择器：直接找包含 label 的表单项下的 input）
	fmt.Println("正在填写中文标题...")
	locatorCh := page.Locator(`.ant-form-item-row:has(label[title="产品标题"]) input.ant-input`).First()
	if err := locatorCh.Fill(product.TitleCh); err != nil {
		return fmt.Errorf("输入中文标题失败: %w", err)
	}
	fmt.Println("中文标题填写完成")

	// 填写英文标题
	fmt.Println("正在填写英文标题...")
	locatorEn := page.Locator(`.ant-form-item-row:has(label[title="英文标题"]) input.ant-input`).First()
	if err := locatorEn.Fill(product.TitleEn); err != nil {
		return fmt.Errorf("输入英文标题失败: %w", err)
	}
	fmt.Println("英文标题填写完成")

	// 上传产品素材图
	fmt.Println("正在上传产品素材图...")
	if err := uploadMaterialImage(page, product.MaterialImg); err != nil {
		return fmt.Errorf("上传产品素材图失败: %w", err)
	}
	fmt.Println("产品素材图上传完成")

	// 清空品牌图片
	fmt.Println("清空品牌图片...")
	if err := clearBianzhongImage(page); err != nil {
		return fmt.Errorf("上传品牌图片失败: %w", err)
	}

	// 设置品牌图片
	if err := setBianzhongImage(page, product.MaterialImg); err != nil {
		return fmt.Errorf("上传品牌图片失败: %w", err)
	}
	fmt.Println("品牌图片上传完成")
	return nil
}

// uploadMaterialImage 上传产品素材图
func uploadMaterialImage(page playwright.Page, imagePath string) error {
	if imagePath == "" {
		return nil
	}

	// 定位"产品素材图"区域中的图片元素
	imgElement := page.Locator(`.ant-form-item-row:has(label[title="产品素材图"]) .img-css`).First()

	// 等待图片元素可见
	fmt.Println("等待产品素材图元素可见...")
	if err := imgElement.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("等待图片元素可见失败: %w", err)
	}

	// 滚动到元素可见位置
	if err := imgElement.ScrollIntoViewIfNeeded(); err != nil {
		return fmt.Errorf("滚动到图片元素失败: %w", err)
	}

	// 步骤1: 点击图片元素，触发下拉菜单显示
	fmt.Println("点击产品素材图...")
	if err := imgElement.Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("点击图片元素失败: %w", err)
	}
	fmt.Println("已点击产品素材图")

	// 等待一下让悬浮菜单显示
	page.WaitForTimeout(500)

	// 步骤2: 等待下拉菜单出现并点击"本地图片"选项
	localMenuItem := page.Locator(".ant-dropdown-menu-item[data-menu-id=\"local\"]")
	if err := localMenuItem.WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("等待本地图片菜单项失败: %w", err)
	}

	// 步骤3: 设置文件选择器监听，然后点击"本地图片"
	fileChooser, err := page.ExpectFileChooser(func() error {
		return localMenuItem.Click(playwright.LocatorClickOptions{
			Timeout: playwright.Float(5000),
		})
	})
	if err != nil {
		return fmt.Errorf("点击本地图片菜单项失败: %w", err)
	}

	// 步骤4: 设置上传文件
	if err := fileChooser.SetFiles(imagePath); err != nil {
		return fmt.Errorf("设置上传文件失败: %w", err)
	}

	fmt.Printf("已上传产品素材图: %s\n", imagePath)

	// 等待图片上传完成
	page.WaitForTimeout(1000)

	return nil
}

// 清空品牌图片
func clearBianzhongImage(page playwright.Page) error {
	// 步骤1: 点击"批量"按钮，触发下拉菜单
	batchBtn := page.Locator("a.img-options-action-btn.ant-dropdown-trigger:has-text(\"批量\")").First()

	fmt.Println("等待批量按钮可见...")
	if err := batchBtn.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("等待批量按钮失败: %w", err)
	}

	// 滚动到元素可见位置
	if err := batchBtn.ScrollIntoViewIfNeeded(); err != nil {
		return fmt.Errorf("滚动到批量按钮失败: %w", err)
	}

	fmt.Println("点击批量按钮...")
	if err := batchBtn.Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("点击批量按钮失败: %w", err)
	}
	fmt.Println("已点击批量按钮")

	// 等待悬浮菜单显示
	page.WaitForTimeout(500)

	// 步骤2: 等待下拉菜单出现并点击"清空图片"选项
	clearMenuItem := page.Locator(".ant-dropdown-menu-item:has-text(\"清空图片\")")
	fmt.Println("等待清空图片菜单项...")
	if err := clearMenuItem.WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("等待清空图片菜单项失败: %w", err)
	}

	fmt.Println("点击清空图片...")
	if err := clearMenuItem.Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("点击清空图片失败: %w", err)
	}
	fmt.Println("已点击清空图片")

	// 步骤3: 等待确认弹窗出现并点击"确定"按钮
	confirmBtn := page.Locator(".ant-modal-content .ant-btn-primary:has-text(\"确 定\")")
	fmt.Println("等待确认弹窗...")
	if err := confirmBtn.WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("等待确认弹窗失败: %w", err)
	}

	fmt.Println("点击确定按钮...")
	if err := confirmBtn.Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("点击确定按钮失败: %w", err)
	}
	fmt.Println("已确认清空图片")

	// 等待弹窗关闭
	page.WaitForTimeout(1000)

	return nil
}

// 设置品牌图片
func setBianzhongImage(page playwright.Page, imagePaths string) error {
	// 获取表格中的所有行
	rows := page.Locator("#skuAttrsInfo tbody tr")
	count, err := rows.Count()
	if err != nil {
		return fmt.Errorf("获取表格行数失败: %w", err)
	}

	fmt.Printf("找到 %d 行品牌数据\n", count)

	// 遍历每一行
	for i := 0; i < count; i++ {
		row := rows.Nth(i)

		// 获取颜色名称（第一列）
		colorCell := row.Locator("td.color-table-cell:first-child")
		colorName, err := colorCell.TextContent()
		if err != nil {
			fmt.Printf("获取第 %d 行颜色名称失败: %v\n", i+1, err)
			continue
		}
		fmt.Printf("第 %d 行颜色: %s\n", i+1, colorName)

		// 获取"选择图片"按钮（第三列）
		selectImgBtn := row.Locator("button:has-text(\"选择图片\")")

		// 滚动到按钮可见位置
		if err := selectImgBtn.ScrollIntoViewIfNeeded(); err != nil {
			fmt.Printf("滚动到第 %d 行按钮失败: %v\n", i+1, err)
			continue
		}

		// 点击"选择图片"按钮
		fmt.Printf("点击第 %d 行的选择图片按钮...\n", i+1)
		if err := selectImgBtn.Click(playwright.LocatorClickOptions{
			Timeout: playwright.Float(5000),
		}); err != nil {
			fmt.Printf("点击第 %d 行选择图片按钮失败: %v\n", i+1, err)
			continue
		}

		// 等待悬浮菜单显示
		page.WaitForTimeout(500)

		// 等待并点击"本地图片"菜单项（只选择可见的那个）
		localMenuItem := page.Locator(".ant-dropdown-menu-item[data-menu-id=\"local\"]:visible").First()
		if err := localMenuItem.WaitFor(playwright.LocatorWaitForOptions{
			Timeout: playwright.Float(5000),
		}); err != nil {
			fmt.Printf("等待第 %d 行本地图片菜单项失败: %v\n", i+1, err)
			continue
		}

		// 设置文件选择器监听，然后点击"本地图片"
		fileChooser, err := page.ExpectFileChooser(func() error {
			return localMenuItem.Click(playwright.LocatorClickOptions{
				Timeout: playwright.Float(5000),
			})
		})
		if err != nil {
			fmt.Printf("点击第 %d 行本地图片菜单项失败: %v\n", i+1, err)
			continue
		}

		// 设置上传文件
		imagePath := getImagePathByColor(colorName)
		if err := fileChooser.SetFiles(imagePath); err != nil {
			fmt.Printf("设置第 %d 行上传文件失败: %v\n", i+1, err)
			continue
		}

		fmt.Printf("第 %d 行 (%s) 图片上传完成\n", i+1, colorName)

		// 等待图片上传完成
		page.WaitForTimeout(1000)
	}

	return nil
}

func getImagePathByColor(color string) []string {
	var result []string

	// 获取 image 目录路径
	imageDir := "image"

	// 遍历 image 目录
	files, err := os.ReadDir(imageDir)
	if err != nil {
		fmt.Printf("读取 image 目录失败: %v\n", err)
		return result
	}

	// 查找文件名包含 color 的文件
	for _, file := range files {
		if !file.IsDir() && strings.Contains(file.Name(), color) {
			fullPath := filepath.Join(imageDir, file.Name())
			result = append(result, fullPath)
		}
	}

	return result
}
