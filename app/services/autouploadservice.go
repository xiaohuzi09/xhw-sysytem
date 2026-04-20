package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// AutoUploadService 自动上传服务
type AutoUploadService struct {
	app              *application.App
	page             playwright.Page
	pw               *playwright.Playwright
	browser          playwright.Browser
	loginConfirmChan chan bool // 用于等待前端确认登录完成
}

// LoginConfig 登录配置
type LoginConfig struct {
	URL             string `json:"url"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	KeepBrowserOpen bool   `json:"keepBrowserOpen"` // 任务完成后是否保持浏览器开启
}

// ProductInfo 商品信息（店小秘平台）
type ProductInfo struct {
	TitleCh     string   `json:"titleCh"`     // 中文标题
	TitleEn     string   `json:"titleEn"`     // 英文标题
	ImagePaths  []string `json:"imagePaths"`  // 本地图片路径列表
	MaterialImg string   `json:"materialImg"` // 产品素材图路径
}

// UploadResult 上传结果
type UploadResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	URL     string `json:"url,omitempty"`
}

// ServiceStartup 服务启动时调用
func (s *AutoUploadService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	s.app = application.Get()
	s.loginConfirmChan = make(chan bool, 1)
	return nil
}

// InitBrowser 初始化浏览器
func (s *AutoUploadService) InitBrowser() error {
	// 启动 Playwright
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("启动 Playwright 失败: %w", err)
	}
	s.pw = pw

	// 启动浏览器
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false), // 非 headless 模式以便观察
	})
	if err != nil {
		return fmt.Errorf("启动浏览器失败: %w", err)
	}
	s.browser = browser

	// 创建新页面
	page, err := browser.NewPage()
	if err != nil {
		return fmt.Errorf("创建页面失败: %w", err)
	}
	s.page = page

	return nil
}

// CloseBrowser 关闭浏览器
func (s *AutoUploadService) CloseBrowser() error {
	if s.page != nil {
		if err := s.page.Close(); err != nil {
			return err
		}
		s.page = nil
	}
	if s.browser != nil {
		if err := s.browser.Close(); err != nil {
			return err
		}
		s.browser = nil
	}
	if s.pw != nil {
		if err := s.pw.Stop(); err != nil {
			return err
		}
		s.pw = nil
	}
	return nil
}

// Login 登录到店小秘平台（支持人工处理验证码）
func (s *AutoUploadService) Login(config LoginConfig) error {
	if s.page == nil {
		return fmt.Errorf("浏览器未初始化，请先调用 InitBrowser")
	}

	// 导航到登录页面
	if _, err := s.page.Goto(config.URL); err != nil {
		return fmt.Errorf("导航到登录页面失败: %w", err)
	}

	// 等待页面加载完成
	if err := s.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		return fmt.Errorf("等待登录页面加载失败: %w", err)
	}

	// 输入用户名
	if err := s.page.Locator("input[name='account']").Fill(config.Username); err != nil {
		return fmt.Errorf("输入用户名失败: %w", err)
	}

	// 输入密码
	if err := s.page.Locator("input[name='password']").Fill(config.Password); err != nil {
		return fmt.Errorf("输入密码失败: %w", err)
	}

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("请在前端界面点击【登录完成】按钮继续")
	fmt.Println("========================================")
	fmt.Println()

	// 等待前端确认登录完成（通过通道阻塞）
	<-s.loginConfirmChan

	// 等待页面稳定
	time.Sleep(2 * time.Second)

	return nil
}

// ConfirmLogin 前端调用此方法确认登录完成
func (s *AutoUploadService) ConfirmLogin() error {
	if s.loginConfirmChan == nil {
		return fmt.Errorf("登录确认通道未初始化")
	}
	select {
	case s.loginConfirmChan <- true:
		return nil
	default:
		return fmt.Errorf("登录确认通道已满")
	}
}

// ClickEditDescription 点击"编辑描述"按钮并等待弹窗出现
func (s *AutoUploadService) ClickEditDescription() error {
	if s.page == nil {
		return fmt.Errorf("浏览器未初始化")
	}

	fmt.Println("正在查找产品描述区域...")

	// 第一步：滚动到产品描述区域并点击遮罩层显示编辑按钮
	descBox := s.page.Locator("#wirelessDescBox").First()

	// 等待描述区域可见
	if err := descBox.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("等待产品描述区域失败: %w", err)
	}

	// 滚动到描述区域
	fmt.Println("滚动到产品描述区域...")
	if err := descBox.ScrollIntoViewIfNeeded(); err != nil {
		return fmt.Errorf("滚动到产品描述区域失败: %w", err)
	}

	// 等待一下确保滚动完成
	s.page.WaitForTimeout(500)

	// 点击遮罩层区域来显示编辑按钮
	shadowDiv := s.page.Locator("#wirelessDescBox .wireless-description-shadow").First()
	fmt.Println("点击遮罩层显示编辑按钮...")
	if err := shadowDiv.Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		// 如果遮罩层点击失败，尝试直接点击描述盒子
		fmt.Println("遮罩层点击失败，尝试点击描述盒子...")
		if err := descBox.Click(playwright.LocatorClickOptions{
			Timeout: playwright.Float(5000),
		}); err != nil {
			return fmt.Errorf("点击产品描述区域失败: %w", err)
		}
	}

	// 等待编辑按钮显示
	s.page.WaitForTimeout(800)

	fmt.Println("正在查找编辑描述按钮...")

	// 第二步：定位并点击"编辑描述"按钮
	editBtn := s.page.Locator("#baiduStatisticsSmtNewEditorEditClickNum button").First()

	// 等待按钮可见
	if err := editBtn.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(1000),
	}); err != nil {
		// 尝试备用选择器
		editBtn = s.page.Locator(".wireless-description-shadow button:has-text(\"编辑描述\")").First()
		if err := editBtn.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1000),
		}); err != nil {
			return fmt.Errorf("等待编辑描述按钮失败: %w", err)
		}
	}

	fmt.Println("点击编辑描述按钮...")

	// 直接点击按钮，不再滚动
	if err := editBtn.Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(10000),
		Force:   playwright.Bool(true), // 强制点击，即使元素被遮挡
	}); err != nil {
		return fmt.Errorf("点击编辑描述按钮失败: %w", err)
	}

	fmt.Println("已点击编辑描述按钮，等待弹窗出现...")

	// 等待弹窗出现
	modalSelectors := []string{
		".ant-modal-wrap",
		".ant-modal-content",
		".tox-dialog",
		"[role=\"dialog\"]",
		".editor-modal",
		".rich-text-editor-modal",
	}

	var modalVisible bool
	for _, selector := range modalSelectors {
		modal := s.page.Locator(selector).First()
		if err := modal.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(5000),
		}); err == nil {
			modalVisible = true
			fmt.Printf("弹窗已出现 (选择器: %s)\n", selector)
			break
		}
	}

	if !modalVisible {
		s.page.WaitForTimeout(2000)
		fmt.Println("已等待弹窗渲染")
	}

	s.page.WaitForTimeout(1000)

	fmt.Println("编辑描述弹窗已准备就绪")
	return nil
}

// ChangeDetailImage 在详情弹窗中更换图片
// 流程：悬停图片 -> 点击"更换图片" -> 点击"本地上传" -> 设置图片
func (s *AutoUploadService) ChangeDetailImage(imagePath string) error {
	if s.page == nil {
		return fmt.Errorf("浏览器未初始化")
	}

	fmt.Println("开始更换详情图片...")

	// 步骤1: 等待弹窗中的图片区域可见
	// 使用多种选择器尝试定位图片模块
	selectors := []string{
		".detail-info-content .item-image-content.has-image",
		".detail-info-content .item-image",
		".detail-info-content .content .item",
		"[class*='detail-info'] img",
	}

	var imageContainer playwright.Locator
	var found bool
	for _, selector := range selectors {
		locator := s.page.Locator(selector).First()
		if err := locator.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(3000),
		}); err == nil {
			imageContainer = locator
			found = true
			fmt.Printf("找到图片容器: %s\n", selector)
			break
		}
	}

	if !found {
		return fmt.Errorf("未找到图片容器元素")
	}

	// 步骤2: 查找"更换图片"链接
	// 尝试多种选择器定位"更换图片"链接
	changeLinkSelectors := []string{
		".detail-info-content .item-image-content a:has-text(\"更换图片\")",
		".detail-info-content a:has-text(\"更换图片\")",
		".item-image-content a:has-text(\"更换图片\")",
		"a:has-text(\"更换图片\")",
	}

	var changeLink playwright.Locator
	var linkFound bool
	for _, selector := range changeLinkSelectors {
		locator := s.page.Locator(selector).First()
		if count, err := locator.Count(); err == nil && count > 0 {
			changeLink = locator
			linkFound = true
			fmt.Printf("找到'更换图片'链接: %s\n", selector)
			break
		}
	}

	if !linkFound {
		// 尝试直接悬停在图片容器上，可能链接是hover时才显示的
		fmt.Println("未找到'更换图片'链接，尝试悬停显示...")
		if err := imageContainer.Hover(playwright.LocatorHoverOptions{
			Timeout: playwright.Float(3000),
		}); err != nil {
			return fmt.Errorf("悬停图片容器失败: %w", err)
		}
		s.page.WaitForTimeout(500)

		// 再次尝试查找链接
		for _, selector := range changeLinkSelectors {
			locator := s.page.Locator(selector).First()
			if count, err := locator.Count(); err == nil && count > 0 {
				if visible, _ := locator.IsVisible(); visible {
					changeLink = locator
					linkFound = true
					fmt.Printf("悬停后找到'更换图片'链接: %s\n", selector)
					break
				}
			}
		}
	}

	if !linkFound {
		return fmt.Errorf("未找到'更换图片'链接")
	}

	// 步骤3: 点击"更换图片"链接
	// 使用 Force 点击，避免滚动循环问题
	fmt.Println("点击'更换图片'...")
	if err := changeLink.Click(playwright.LocatorClickOptions{
		Force:   playwright.Bool(true),
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("点击'更换图片'失败: %w", err)
	}

	// 等待悬浮菜单显示
	s.page.WaitForTimeout(800)

	// 步骤4: 点击"本地上传"
	localUploadSelectors := []string{
		".ant-dropdown-menu-item:has-text(\"本地上传\")",
		".ant-dropdown-menu-item:has-text(\"本地图片\")",
		"[data-menu-id=\"local\"]",
		".ant-dropdown-menu-item:has-text(\"上传\")",
	}

	var localUploadBtn playwright.Locator
	var uploadFound bool
	for _, selector := range localUploadSelectors {
		locator := s.page.Locator(selector).First()
		if err := locator.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(3000),
		}); err == nil {
			localUploadBtn = locator
			uploadFound = true
			fmt.Printf("找到'本地上传'菜单项: %s\n", selector)
			break
		}
	}

	if !uploadFound {
		return fmt.Errorf("未找到'本地上传'菜单项")
	}

	// 设置文件选择器监听，然后点击"本地上传"
	fmt.Println("点击'本地上传'并等待文件选择器...")
	fileChooser, err := s.page.ExpectFileChooser(func() error {
		return localUploadBtn.Click(playwright.LocatorClickOptions{
			Timeout: playwright.Float(5000),
		})
	})
	if err != nil {
		return fmt.Errorf("点击'本地上传'失败: %w", err)
	}

	// 步骤5: 设置上传文件
	fmt.Printf("设置上传文件: %s\n", imagePath)
	if err := fileChooser.SetFiles(imagePath); err != nil {
		return fmt.Errorf("设置上传文件失败: %w", err)
	}

	fmt.Println("图片更换完成")

	// 等待图片上传完成
	s.page.WaitForTimeout(1500)

	// 步骤6: 点击保存按钮
	fmt.Println("点击保存按钮...")
	saveBtnSelectors := []string{
		".top-header .btn-orange:has-text(\"保存\")",
		".top-header button:has-text(\"保存\")",
		".title-right .btn-orange",
		"button.btn-orange:has-text(\"保存\")",
	}

	var saveBtn playwright.Locator
	var btnFound bool
	for _, selector := range saveBtnSelectors {
		locator := s.page.Locator(selector).First()
		if err := locator.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(3000),
		}); err == nil {
			saveBtn = locator
			btnFound = true
			fmt.Printf("找到保存按钮: %s\n", selector)
			break
		}
	}

	if !btnFound {
		return fmt.Errorf("未找到保存按钮")
	}

	if err := saveBtn.Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("点击保存按钮失败: %w", err)
	}

	fmt.Println("已点击保存按钮")

	// 等待保存完成
	s.page.WaitForTimeout(1000)

	return nil
}

// PublishProduct 悬浮在发布按钮上，然后点击立即发布
func (s *AutoUploadService) PublishProduct() error {
	if s.page == nil {
		return fmt.Errorf("浏览器未初始化")
	}

	fmt.Println("开始发布产品...")

	// 步骤1: 查找发布按钮
	publishBtnSelectors := []string{
		"button.btn-green:has-text(\"发布\")",
		".btn-green:has-text(\"发布\")",
		"button.ant-btn.btn-green",
		"button:has-text(\"发布\")",
	}

	var publishBtn playwright.Locator
	var btnFound bool
	for _, selector := range publishBtnSelectors {
		locator := s.page.Locator(selector).First()
		if err := locator.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(5000),
		}); err == nil {
			publishBtn = locator
			btnFound = true
			fmt.Printf("找到发布按钮: %s\n", selector)
			break
		}
	}

	if !btnFound {
		return fmt.Errorf("未找到发布按钮")
	}

	// 步骤2: 悬停在发布按钮上，显示下拉菜单
	fmt.Println("悬停在发布按钮上...")
	if err := publishBtn.Hover(playwright.LocatorHoverOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("悬停发布按钮失败: %w", err)
	}

	// 等待下拉菜单显示
	s.page.WaitForTimeout(500)

	// 步骤3: 点击"立即发布"菜单项
	fmt.Println("查找'立即发布'菜单项...")
	publishNowSelectors := []string{
		".ant-dropdown-menu-item:has-text(\"立即发布\")",
		".ant-dropdown-menu-item:has-text(\"发布\")",
		"[data-menu-id]:has-text(\"立即发布\")",
		".ant-dropdown-menu li:has-text(\"立即发布\")",
	}

	var publishNowBtn playwright.Locator
	var publishFound bool
	for _, selector := range publishNowSelectors {
		locator := s.page.Locator(selector).First()
		if err := locator.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(3000),
		}); err == nil {
			publishNowBtn = locator
			publishFound = true
			fmt.Printf("找到'立即发布'菜单项: %s\n", selector)
			break
		}
	}

	if !publishFound {
		return fmt.Errorf("未找到'立即发布'菜单项")
	}

	fmt.Println("点击'立即发布'...")
	if err := publishNowBtn.Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("点击'立即发布'失败: %w", err)
	}

	fmt.Println("已点击立即发布")

	// 等待发布完成
	s.page.WaitForTimeout(1000)

	return nil
}

// ClickFirstImageItem 点击编辑器中的第一个图片项
func (s *AutoUploadService) ClickFirstImageItem() error {
	if s.page == nil {
		return fmt.Errorf("浏览器未初始化")
	}

	fmt.Println("正在查找第一个图片项...")

	// 等待图片列表容器加载
	container := s.page.Locator(".using-modules-content.sortable-container").First()
	if err := container.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("等待图片列表容器失败: %w", err)
	}

	// 查找第一个图片项（data-idx="0" 表示第一个）
	firstItem := s.page.Locator(".using-item.sortable-item[data-idx=\"0\"]").First()

	// 等待第一个图片项可见
	if err := firstItem.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("等待第一个图片项失败: %w", err)
	}

	// 滚动到第一个图片项
	if err := firstItem.ScrollIntoViewIfNeeded(); err != nil {
		return fmt.Errorf("滚动到第一个图片项失败: %w", err)
	}

	fmt.Println("点击第一个图片项...")

	// 点击图片项
	if err := firstItem.Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("点击第一个图片项失败: %w", err)
	}

	fmt.Println("已点击第一个图片项")
	return nil
}

// NavigateToProductPage 导航到产品编辑页面
func (s *AutoUploadService) NavigateToProductPage(productID string) error {
	if s.page == nil {
		return fmt.Errorf("浏览器未初始化")
	}

	url := "https://www.dianxiaomi.com/web/popTemu/quoteEdit?id=154449843077440818"
	if _, err := s.page.Goto(url); err != nil {
		return fmt.Errorf("导航到产品页面失败: %w", err)
	}

	// 等待页面加载完成
	if err := s.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		return fmt.Errorf("等待产品页面加载失败: %w", err)
	}

	return nil
}

// FillProductInfo 填写产品信息（店小秘平台）
// productID 是素材编号，用于从合成目录查找图片
func (s *AutoUploadService) FillProductInfo(productID string, product ProductInfo) error {
	// 填写中文标题
	fmt.Println("正在填写中文标题...")
	locatorCh := s.page.Locator(`.ant-form-item-row:has(label[title="产品标题"]) input.ant-input`).First()
	if err := locatorCh.Fill(product.TitleCh); err != nil {
		return fmt.Errorf("输入中文标题失败: %w", err)
	}
	fmt.Println("中文标题填写完成")

	// 填写英文标题
	fmt.Println("正在填写英文标题...")
	locatorEn := s.page.Locator(`.ant-form-item-row:has(label[title="英文标题"]) input.ant-input`).First()
	if err := locatorEn.Fill(product.TitleEn); err != nil {
		return fmt.Errorf("输入英文标题失败: %w", err)
	}
	fmt.Println("英文标题填写完成")

	// 从合成目录查找素材图（根据素材编号）
	fmt.Printf("正在从合成目录查找素材编号 %s 的图片...\n", productID)
	materialImagePath, err := s.findMaterialImageByCode(productID)
	if err != nil {
		return fmt.Errorf("查找素材图片失败: %w", err)
	}
	fmt.Printf("找到素材图片: %s\n", materialImagePath)

	// 上传产品素材图
	fmt.Println("正在上传产品素材图...")
	if err := s.uploadMaterialImage(materialImagePath); err != nil {
		return fmt.Errorf("上传产品素材图失败: %w", err)
	}
	fmt.Println("产品素材图上传完成")

	// 清空品牌图片
	fmt.Println("清空品牌图片...")
	if err := s.clearBianzhongImage(); err != nil {
		return fmt.Errorf("清空品牌图片失败: %w", err)
	}

	// 从合成目录查找所有匹配素材编号的图片作为品牌图片
	fmt.Printf("正在查找素材编号 %s 的所有品牌图片...\n", productID)
	brandImages, err := s.findImagesByMaterialCode(productID)
	if err != nil {
		return fmt.Errorf("查找品牌图片失败: %w", err)
	}
	fmt.Printf("找到 %d 张品牌图片\n", len(brandImages))

	// 设置品牌图片
	fmt.Println("正在设置品牌图片...")
	if err := s.setBianzhongImage(brandImages); err != nil {
		return fmt.Errorf("设置品牌图片失败: %w", err)
	}
	fmt.Println("品牌图片设置完成")

	// 点击"编辑描述"按钮
	if err := s.ClickEditDescription(); err != nil {
		return fmt.Errorf("点击编辑描述按钮失败: %w", err)
	}

	fmt.Println("正在点击第一个图片项...")
	if err := s.ClickFirstImageItem(); err != nil {
		return fmt.Errorf("点击第一个图片项失败: %w", err)
	}

	if err := s.ChangeDetailImage(materialImagePath); err != nil {
		return fmt.Errorf("更换详情图片失败: %w", err)
	}

	fmt.Println("详情图片更换完成")

	err = s.PublishProduct()
	if err != nil {
		log.Printf("发布失败: %v", err)
	}

	return nil
}

// uploadMaterialImage 上传产品素材图
func (s *AutoUploadService) uploadMaterialImage(imagePath string) error {
	if imagePath == "" {
		return nil
	}

	// 定位"产品素材图"区域中的图片元素
	imgElement := s.page.Locator(`.ant-form-item-row:has(label[title="产品素材图"]) .img-css`).First()

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
	s.page.WaitForTimeout(500)

	// 步骤2: 等待下拉菜单出现并点击"本地图片"选项
	localMenuItem := s.page.Locator(".ant-dropdown-menu-item[data-menu-id=\"local\"]")
	if err := localMenuItem.WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("等待本地图片菜单项失败: %w", err)
	}

	// 步骤3: 设置文件选择器监听，然后点击"本地图片"
	fileChooser, err := s.page.ExpectFileChooser(func() error {
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
	s.page.WaitForTimeout(1000)

	return nil
}

// clearBianzhongImage 清空品牌图片
func (s *AutoUploadService) clearBianzhongImage() error {
	// 先等待页面完全加载
	fmt.Println("等待页面稳定...")
	s.page.WaitForTimeout(2000)

	// 步骤1: 点击"批量"按钮，触发下拉菜单
	// 使用更通用的选择器
	batchBtn := s.page.Locator(".img-options-action-btn:has-text(\"批量\")").First()

	fmt.Println("等待批量按钮可见...")
	// 增加超时时间到 30 秒
	if err := batchBtn.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(30000),
	}); err != nil {
		// 如果找不到批量按钮，可能是页面结构不同，尝试备用选择器
		fmt.Println("使用备用选择器查找批量按钮...")
		batchBtn = s.page.Locator("button:has-text(\"批量\")").First()
		if err := batchBtn.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(30000),
		}); err != nil {
			return fmt.Errorf("等待批量按钮失败: %w", err)
		}
	}

	// 滚动到元素可见位置
	fmt.Println("滚动到批量按钮...")
	if err := batchBtn.ScrollIntoViewIfNeeded(); err != nil {
		return fmt.Errorf("滚动到批量按钮失败: %w", err)
	}

	fmt.Println("点击批量按钮...")
	if err := batchBtn.Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("点击批量按钮失败: %w", err)
	}
	fmt.Println("已点击批量按钮")

	// 等待悬浮菜单显示
	s.page.WaitForTimeout(1000)

	// 步骤2: 等待下拉菜单出现并点击"清空图片"选项
	clearMenuItem := s.page.Locator(".ant-dropdown-menu-item:has-text(\"清空图片\")").First()
	fmt.Println("等待清空图片菜单项...")
	if err := clearMenuItem.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("等待清空图片菜单项失败: %w", err)
	}

	fmt.Println("点击清空图片...")
	if err := clearMenuItem.Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("点击清空图片失败: %w", err)
	}
	fmt.Println("已点击清空图片")

	// 等待弹窗出现
	s.page.WaitForTimeout(1000)

	// 步骤3: 等待确认弹窗出现并点击"确定"按钮
	confirmBtn := s.page.Locator(".ant-modal-content .ant-btn-primary:has-text(\"确 定\")").First()
	fmt.Println("等待确认弹窗...")
	if err := confirmBtn.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	}); err != nil {
		// 尝试其他可能的选择器
		confirmBtn = s.page.Locator(".ant-modal-wrap .ant-btn-primary").First()
		if err := confirmBtn.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(10000),
		}); err != nil {
			return fmt.Errorf("等待确认弹窗失败: %w", err)
		}
	}

	fmt.Println("点击确定按钮...")
	if err := confirmBtn.Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(10000),
	}); err != nil {
		return fmt.Errorf("点击确定按钮失败: %w", err)
	}
	fmt.Println("已确认清空图片")

	// 等待弹窗关闭
	s.page.WaitForTimeout(2000)

	return nil
}

// setBianzhongImage 设置品牌图片（根据颜色匹配）
func (s *AutoUploadService) setBianzhongImage(imagePaths []string) error {
	// 获取表格中的所有行
	rows := s.page.Locator("#skuAttrsInfo tbody tr")
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
		s.page.WaitForTimeout(500)

		// 等待并点击"本地图片"菜单项（只选择可见的那个）
		localMenuItem := s.page.Locator(".ant-dropdown-menu-item[data-menu-id=\"local\"]:visible").First()
		if err := localMenuItem.WaitFor(playwright.LocatorWaitForOptions{
			Timeout: playwright.Float(5000),
		}); err != nil {
			fmt.Printf("等待第 %d 行本地图片菜单项失败: %v\n", i+1, err)
			continue
		}

		// 设置文件选择器监听，然后点击"本地图片"
		fileChooser, err := s.page.ExpectFileChooser(func() error {
			return localMenuItem.Click(playwright.LocatorClickOptions{
				Timeout: playwright.Float(5000),
			})
		})
		if err != nil {
			fmt.Printf("点击第 %d 行本地图片菜单项失败: %v\n", i+1, err)
			continue
		}

		// 根据颜色查找匹配的图片
		matchedImages := s.getImagePathsByColor(colorName, imagePaths)
		if len(matchedImages) == 0 {
			fmt.Printf("第 %d 行 (%s) 未找到匹配的图片\n", i+1, colorName)
			continue
		}

		// 设置上传文件（支持多图）
		if err := fileChooser.SetFiles(matchedImages); err != nil {
			fmt.Printf("设置第 %d 行上传文件失败: %v\n", i+1, err)
			continue
		}

		fmt.Printf("第 %d 行 (%s) 图片上传完成，共 %d 张\n", i+1, colorName, len(matchedImages))

		// 等待图片上传完成
		s.page.WaitForTimeout(1000)
	}

	return nil
}

// getImagePathsByColor 根据颜色名称从图片路径列表中筛选匹配的图片
func (s *AutoUploadService) getImagePathsByColor(color string, imagePaths []string) []string {
	var result []string
	colorLower := strings.ToLower(color)

	for _, path := range imagePaths {
		filename := strings.ToLower(filepath.Base(path))
		if strings.Contains(filename, colorLower) {
			result = append(result, path)
		}
	}

	return result
}

// findMaterialImageByCode 从合成目录中根据素材编号查找匹配的图片
// 文件名格式: 任务批号-模板名称-颜色-素材编号-时间戳.png
// 返回匹配素材编号的第一个图片路径
func (s *AutoUploadService) findMaterialImageByCode(materialCode string) (string, error) {
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取当前目录失败: %w", err)
	}

	// 合成目录路径
	combineDir := filepath.Join(currentDir, "合成")

	// 读取合成目录
	entries, err := os.ReadDir(combineDir)
	if err != nil {
		return "", fmt.Errorf("读取合成目录失败: %w", err)
	}

	// 遍历文件，查找匹配的素材编号
	// 文件名格式: 任务批号-模板名称-颜色-素材编号-时间戳.png
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		// 去掉扩展名
		ext := filepath.Ext(filename)
		nameWithoutExt := strings.TrimSuffix(filename, ext)

		// 按"-"分割文件名
		parts := strings.Split(nameWithoutExt, "-")
		// 格式: 任务批号-模板名称-颜色-素材编号-时间戳
		// 素材编号在倒数第二个位置（索引 len(parts)-2）
		if len(parts) >= 2 {
			fileMaterialCode := parts[len(parts)-2]
			if fileMaterialCode == materialCode {
				return filepath.Join(combineDir, filename), nil
			}
		}
	}

	return "", fmt.Errorf("未找到素材编号 %s 对应的图片", materialCode)
}

// findImagesByMaterialCode 从合成目录中查找所有匹配素材编号的图片
// 文件名格式: 任务批号-模板名称-颜色-素材编号-时间戳.png
func (s *AutoUploadService) findImagesByMaterialCode(materialCode string) ([]string, error) {
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("获取当前目录失败: %w", err)
	}

	// 合成目录路径
	combineDir := filepath.Join(currentDir, "合成")

	// 读取合成目录
	entries, err := os.ReadDir(combineDir)
	if err != nil {
		return nil, fmt.Errorf("读取合成目录失败: %w", err)
	}

	var matchedImages []string

	// 遍历文件，查找匹配的素材编号
	// 文件名格式: 任务批号-模板名称-颜色-素材编号-时间戳.png
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		ext := filepath.Ext(filename)
		nameWithoutExt := strings.TrimSuffix(filename, ext)

		// 按"-"分割文件名
		parts := strings.Split(nameWithoutExt, "-")
		// 格式: 任务批号-模板名称-颜色-素材编号-时间戳
		// 素材编号在倒数第二个位置（索引 len(parts)-2）
		if len(parts) >= 2 {
			fileMaterialCode := parts[len(parts)-2]
			if fileMaterialCode == materialCode {
				matchedImages = append(matchedImages, filepath.Join(combineDir, filename))
			}
		}
	}

	if len(matchedImages) == 0 {
		return nil, fmt.Errorf("未找到素材编号 %s 对应的图片", materialCode)
	}

	return matchedImages, nil
}

// UploadProductDianxiaomi 店小秘平台自动上传商品
// 注意：此方法不会自动关闭浏览器，上传完成后需要手动调用 CloseBrowser() 关闭
func (s *AutoUploadService) UploadProductDianxiaomi(loginConfig LoginConfig, productID string, product ProductInfo) (*UploadResult, error) {
	result := &UploadResult{
		Success: false,
	}

	// 1. 初始化浏览器
	if err := s.InitBrowser(); err != nil {
		result.Message = fmt.Sprintf("初始化浏览器失败: %v", err)
		return result, err
	}
	// 注意：不再自动关闭浏览器，需要手动调用 CloseBrowser()

	// 2. 调用登录方法
	if err := s.Login(loginConfig); err != nil {
		result.Message = fmt.Sprintf("登录失败: %v", err)
		return result, err
	}

	// 3. 导航到产品编辑页面
	if err := s.NavigateToProductPage(productID); err != nil {
		result.Message = fmt.Sprintf("导航到产品页面失败: %v", err)
		return result, err
	}

	// 4. 填写产品信息
	if err := s.FillProductInfo(productID, product); err != nil {
		result.Message = fmt.Sprintf("填写产品信息失败: %v", err)
		return result, err
	}

	result.Success = true
	result.Message = "商品信息填写完成，浏览器保持打开状态，请手动确认提交"
	return result, nil
}

// UploadProductsDianxiaomi 批量上传多个商品到店小秘
// 注意：此方法不会自动关闭浏览器，上传完成后需要手动调用 CloseBrowser() 关闭
func (s *AutoUploadService) UploadProductsDianxiaomi(loginConfig LoginConfig, products map[string]ProductInfo) ([]*UploadResult, error) {
	results := make([]*UploadResult, 0, len(products))

	// 1. 初始化浏览器并登录（只执行一次）
	if err := s.InitBrowser(); err != nil {
		result := &UploadResult{
			Success: false,
			Message: fmt.Sprintf("初始化浏览器失败: %v", err),
		}
		return append(results, result), err
	}
	// 注意：不再自动关闭浏览器，需要手动调用 CloseBrowser()

	if err := s.Login(loginConfig); err != nil {
		result := &UploadResult{
			Success: false,
			Message: fmt.Sprintf("登录失败: %v", err),
		}
		return append(results, result), err
	}

	// 2. 逐个处理商品
	i := 0
	for productID, product := range products {
		fmt.Printf("\n===== 处理第 %d 个商品: %s =====\n", i+1, productID)

		// 导航到产品页面
		if err := s.NavigateToProductPage(productID); err != nil {
			results = append(results, &UploadResult{
				Success: false,
				Message: fmt.Sprintf("商品 %s 导航失败: %v", productID, err),
			})
			continue
		}

		// 填写产品信息
		if err := s.FillProductInfo(productID, product); err != nil {
			results = append(results, &UploadResult{
				Success: false,
				Message: fmt.Sprintf("商品 %s 填写信息失败: %v", productID, err),
			})
			continue
		}

		results = append(results, &UploadResult{
			Success: true,
			Message: fmt.Sprintf("商品 %s 信息填写完成", productID),
		})

		// 每个商品处理后等待一段时间
		if i < len(products)-1 {
			time.Sleep(3 * time.Second)
		}
		i++
	}

	fmt.Println("\n所有商品处理完成")

	// 根据配置决定是否关闭浏览器
	if !loginConfig.KeepBrowserOpen {
		fmt.Println("正在关闭浏览器...")
		if err := s.CloseBrowser(); err != nil {
			fmt.Printf("关闭浏览器失败: %v\n", err)
		} else {
			fmt.Println("浏览器已关闭")
		}
	} else {
		fmt.Println("浏览器保持打开状态")
	}

	return results, nil
}
