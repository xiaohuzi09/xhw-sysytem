package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	xdraw "golang.org/x/image/draw"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type ImageService struct {
	app *application.App
}

// ImageTemplate 图片模板结构
type ImageTemplate struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	Scale        float64   `json:"scale"`
	ImagePath    string    `json:"imagePath"`
	OffsetTop    int       `json:"offsetTop"`
	OffsetRight  int       `json:"offsetRight"`
	OffsetBottom int       `json:"offsetBottom"`
	OffsetLeft   int       `json:"offsetLeft"`
	CreatedAt    time.Time `json:"createdAt"`
}

var templates []ImageTemplate

// SelectImage 选择图片文件
func (s *ImageService) SelectImage() (string, error) {
	if s.app == nil {
		return "", fmt.Errorf("应用未初始化")
	}

	result, err := s.app.Dialog.OpenFile().
		SetTitle("选择图片").
		AddFilter("图片文件", "*.jpg;*.jpeg;*.png;*.gif;*.bmp;*.webp").
		PromptForSingleSelection()

	if err != nil {
		return "", err
	}

	return result, nil
}

// SelectImages 选择多张图片文件
func (s *ImageService) SelectImages() ([]string, error) {
	if s.app == nil {
		return nil, fmt.Errorf("应用未初始化")
	}

	result, err := s.app.Dialog.OpenFile().
		SetTitle("选择图片").
		AddFilter("图片文件", "*.jpg;*.jpeg;*.png;*.gif;*.bmp;*.webp").
		PromptForMultipleSelection()

	if err != nil {
		return nil, err
	}

	return result, nil
}

// SaveImageToCurrentFolder 保存图片到当前文件夹
func (s *ImageService) SaveImageToCurrentFolder(sourcePath string) (string, error) {
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取当前目录失败: %v", err)
	}

	// 创建 images 子目录
	imagesDir := filepath.Join(currentDir, "images")
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %v", err)
	}

	// 生成新文件名
	fileName := filepath.Base(sourcePath)
	destPath := filepath.Join(imagesDir, fileName)

	// 如果文件已存在，添加时间戳
	if _, err := os.Stat(destPath); err == nil {
		ext := filepath.Ext(fileName)
		nameWithoutExt := fileName[:len(fileName)-len(ext)]
		timestamp := time.Now().Format("20060102_150405")
		fileName = fmt.Sprintf("%s_%s%s", nameWithoutExt, timestamp, ext)
		destPath = filepath.Join(imagesDir, fileName)
	}

	// 复制文件
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("打开源文件失败: %v", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return "", fmt.Errorf("复制文件失败: %v", err)
	}

	return destPath, nil
}

// AddTemplate 添加图片模板
func (s *ImageService) AddTemplate(name string, width, height int, scale float64, imagePath string,
	offsetTop, offsetRight, offsetBottom, offsetLeft int,
) (string, error) {
	if name == "" {
		return "", fmt.Errorf("模板名称不能为空")
	}

	if width <= 0 || height <= 0 {
		return "", fmt.Errorf("宽度和高度必须大于0")
	}

	if scale <= 0 {
		return "", fmt.Errorf("缩放比例必须大于0")
	}

	// 打印图片路径
	fmt.Println("imagePath", imagePath)
	template := ImageTemplate{
		ID:           fmt.Sprintf("%d", time.Now().UnixNano()),
		Name:         name,
		Width:        width,
		Height:       height,
		Scale:        scale,
		ImagePath:    imagePath,
		OffsetTop:    offsetTop,
		OffsetRight:  offsetRight,
		OffsetBottom: offsetBottom,
		OffsetLeft:   offsetLeft,
		CreatedAt:    time.Now(),
	}

	templates = append(templates, template)

	// 保存到文件
	if err := s.saveTemplates(); err != nil {
		return "", fmt.Errorf("保存模板失败: %v", err)
	}

	return template.ID, nil
}

// GetTemplates 获取所有模板
func (s *ImageService) GetTemplates() ([]ImageTemplate, error) {
	return templates, nil
}

// DeleteTemplate 删除模板
func (s *ImageService) DeleteTemplate(id string) error {
	for i, template := range templates {
		if template.ID == id {
			templates = append(templates[:i], templates[i+1:]...)
			return s.saveTemplates()
		}
	}
	return fmt.Errorf("模板不存在")
}

// saveTemplates 保存模板到文件
func (s *ImageService) saveTemplates() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	templatesFile := filepath.Join(currentDir, "templates.json")
	data, err := json.MarshalIndent(templates, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(templatesFile, data, 0644)
}

// loadTemplates 从文件加载模板
func (s *ImageService) loadTemplates() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	templatesFile := filepath.Join(currentDir, "templates.json")
	data, err := os.ReadFile(templatesFile)
	if err != nil {
		if os.IsNotExist(err) {
			templates = []ImageTemplate{}
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &templates)
}

// findTemplateByID 根据 ID 查找模板
func findTemplateByID(id string) (*ImageTemplate, error) {
	for i := range templates {
		if templates[i].ID == id {
			return &templates[i], nil
		}
	}
	return nil, fmt.Errorf("模板不存在")
}

// decodeImage 加载图片文件
func decodeImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开图片失败: %v", err)
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		img, err := jpeg.Decode(file)
		if err != nil {
			return nil, fmt.Errorf("解码 JPEG 失败: %v", err)
		}
		return img, nil
	case ".png":
		img, err := png.Decode(file)
		if err != nil {
			return nil, fmt.Errorf("解码 PNG 失败: %v", err)
		}
		return img, nil
	default:
		// 尝试自动检测格式
		img, _, err := image.Decode(file)
		if err != nil {
			return nil, fmt.Errorf("解码图片失败: %v", err)
		}
		return img, nil
	}
}

// trimImage 裁剪图片四周的空白区域（白色或透明），只保留有效像素
func trimImage(img image.Image) image.Image {
	bounds := img.Bounds()
	minX, minY, maxX, maxY := bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y

	// 判断像素是否为空白（白色或透明）
	isBlank := func(x, y int) bool {
		r, g, b, a := img.At(x, y).RGBA()
		// 如果是完全透明，认为是空白
		if a == 0 {
			return true
		}
		// 如果是白色（或接近白色，阈值设为 65000，约等于 255*255 的 99%），认为是空白
		// RGBA() 返回的是 0-65535 范围的值
		if a >= 65000 && r >= 65000 && g >= 65000 && b >= 65000 {
			return true
		}
		return false
	}

	// 从上边找非空白行
	top := minY
	for top < maxY {
		blankRow := true
		for x := minX; x < maxX; x++ {
			if !isBlank(x, top) {
				blankRow = false
				break
			}
		}
		if !blankRow {
			break
		}
		top++
	}

	// 从下边找非空白行
	bottom := maxY - 1
	for bottom >= top {
		blankRow := true
		for x := minX; x < maxX; x++ {
			if !isBlank(x, bottom) {
				blankRow = false
				break
			}
		}
		if !blankRow {
			break
		}
		bottom--
	}

	// 从左边找非空白列
	left := minX
	for left < maxX {
		blankCol := true
		for y := top; y <= bottom; y++ {
			if !isBlank(left, y) {
				blankCol = false
				break
			}
		}
		if !blankCol {
			break
		}
		left++
	}

	// 从右边找非空白列
	right := maxX - 1
	for right >= left {
		blankCol := true
		for y := top; y <= bottom; y++ {
			if !isBlank(right, y) {
				blankCol = false
				break
			}
		}
		if !blankCol {
			break
		}
		right--
	}

	// 如果没有有效区域，返回原图
	if top > bottom || left > right {
		return img
	}

	// 裁剪出有效区域
	croppedBounds := image.Rect(left, top, right+1, bottom+1)
	cropped := image.NewRGBA(croppedBounds)
	draw.Draw(cropped, croppedBounds, img, image.Point{X: left, Y: top}, draw.Src)

	return cropped
}

// compositeImage 将源图片放入模板中间的框里
func compositeImage(templateImg image.Image, src image.Image, rectWidth, rectHeight int,
	offsetTop, offsetRight, offsetBottom, offsetLeft int,
) *image.RGBA {
	tmplBounds := templateImg.Bounds()
	out := image.NewRGBA(tmplBounds)

	// 先绘制模板底图，然后在框区域内把源图片贴上去
	draw.Draw(out, tmplBounds, templateImg, tmplBounds.Min, draw.Src)

	tw := tmplBounds.Dx()
	th := tmplBounds.Dy()
	if rectWidth <= 0 || rectHeight <= 0 || tw <= 0 || th <= 0 {
		return out
	}

	// 模板框区域，默认居中，再叠加偏移量
	rectX := (tw-rectWidth)/2 + offsetLeft - offsetRight
	rectY := (th-rectHeight)/2 + offsetTop - offsetBottom

	srcBounds := src.Bounds()
	iw := srcBounds.Dx()
	ih := srcBounds.Dy()
	if iw <= 0 || ih <= 0 {
		return out
	}

	// contain 模式缩放：整张图片缩放到“完全放入”框内（可能留边，不裁剪）
	scaleX := float64(rectWidth) / float64(iw)
	scaleY := float64(rectHeight) / float64(ih)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	scaledW := int(float64(iw) * scale)
	scaledH := int(float64(ih) * scale)
	if scaledW <= 0 || scaledH <= 0 {
		return out
	}

	// 使用高质量插值算法进行缩放
	scaledImg := image.NewRGBA(image.Rect(0, 0, scaledW, scaledH))
	xdraw.CatmullRom.Scale(scaledImg, scaledImg.Bounds(), src, srcBounds, xdraw.Over, nil)

	// 让缩放后的图片在框中居中
	offsetX := rectX + (rectWidth-scaledW)/2
	offsetY := rectY + (rectHeight-scaledH)/2

	dstRect := image.Rect(
		tmplBounds.Min.X+offsetX,
		tmplBounds.Min.Y+offsetY,
		tmplBounds.Min.X+offsetX+scaledW,
		tmplBounds.Min.Y+offsetY+scaledH,
	)

	// 带 alpha 的覆盖：透明区域会透出底下的模板
	draw.Draw(out, dstRect, scaledImg, image.Point{}, draw.Over)

	return out
}

// TemplateInfo 用于接收前端传递的模板信息
type TemplateInfo struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	Scale     float64 `json:"scale"`
	ImagePath string  `json:"imagePath"`
	URL       string  `json:"url"`
	OffsetX   int     `json:"offset_x"`
	OffsetY   int     `json:"offset_y"`
}

// MaterialInfo 用于接收前端传递的素材信息
type MaterialInfo struct {
	URL  string `json:"url"`  // 素材图片 URL
	Code string `json:"code"` // 素材编号
}

// CombineImagesWithTemplates 使用模板信息合成多张图片
// 返回本次合成结果所在的目录路径
func (s *ImageService) CombineImagesWithTemplates(templateInfos []TemplateInfo, materials []MaterialInfo) (string, error) {
	if len(templateInfos) == 0 {
		return "", fmt.Errorf("请选择至少一个模板")
	}
	if len(materials) == 0 {
		return "", fmt.Errorf("请选择至少一张图片")
	}

	// 合成目录: 当前目录/合成
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取当前目录失败: %v", err)
	}

	combineDir := filepath.Join(currentDir, "合成")
	if err := os.MkdirAll(combineDir, 0755); err != nil {
		return "", fmt.Errorf("创建合成目录失败: %v", err)
	}

	// 生成批次时间戳前缀，用于区分同一次合成的文件
	batchPrefix := time.Now().Format("20060102_150405")

	for _, templateInfo := range templateInfos {
		// 确定图片路径：优先使用 URL，其次使用 ImagePath
		imagePath := templateInfo.URL
		if imagePath == "" {
			imagePath = templateInfo.ImagePath
		}

		// 加载模板底图
		var templateImg image.Image
		if strings.HasPrefix(imagePath, "http://") || strings.HasPrefix(imagePath, "https://") {
			// URL，直接下载
			templateImg, err = downloadImage(imagePath)
			if err != nil {
				return "", fmt.Errorf("下载模板图片失败 (%s): %v", imagePath, err)
			}
		} else {
			// 本地路径
			templateImg, err = decodeImage(imagePath)
			if err != nil {
				return "", fmt.Errorf("加载模板图片失败 (%s): %v", imagePath, err)
			}
		}

		for _, material := range materials {
			// 加载源图片，支持 URL 和本地路径
			var srcImg image.Image
			if strings.HasPrefix(material.URL, "http://") || strings.HasPrefix(material.URL, "https://") {
				// URL，下载图片
				srcImg, err = downloadImage(material.URL)
				if err != nil {
					return "", fmt.Errorf("加载源图片失败 (%s): %v", material.URL, err)
				}
			} else {
				// 本地路径
				srcImg, err = decodeImage(material.URL)
				if err != nil {
					return "", fmt.Errorf("加载源图片失败 (%s): %v", material.URL, err)
				}
			}

			// 裁剪素材图片四周的空白区域
			trimmedSrcImg := trimImage(srcImg)

			// 计算偏移量：offset_x 是水平偏移，offset_y 是垂直偏移
			offsetLeft := templateInfo.OffsetX
			offsetRight := 0
			offsetTop := templateInfo.OffsetY
			offsetBottom := 0

			outImg := compositeImage(
				templateImg,
				trimmedSrcImg,
				templateInfo.Width,
				templateInfo.Height,
				offsetTop,
				offsetRight,
				offsetBottom,
				offsetLeft,
			)

			// 文件名增加模板名称前缀和时间戳，避免多模板时覆盖
			templateName := strings.TrimSpace(templateInfo.Name)
			if templateName == "" {
				templateName = templateInfo.ID
			}
			// 简单清理模板名中的路径分隔符
			templateName = strings.ReplaceAll(templateName, string(os.PathSeparator), "_")

			// 素材编号，如果没有则使用序号
			materialCode := strings.TrimSpace(material.Code)
			if materialCode == "" {
				materialCode = "unknown"
			}

			// 生成时间戳确保文件名唯一
			timestamp := time.Now().UnixNano()

			// 生成文件名：任务批号-模版名称-素材编号-时间戳.png
			outName := fmt.Sprintf("%s-%s-%s-%d.png", batchPrefix, templateName, materialCode, timestamp)
			outPath := filepath.Join(combineDir, outName)

			outFile, err := os.Create(outPath)
			if err != nil {
				return "", fmt.Errorf("创建合成文件失败 (%s): %v", outPath, err)
			}

			if err := png.Encode(outFile, outImg); err != nil {
				outFile.Close()
				return "", fmt.Errorf("写入合成图片失败 (%s): %v", outPath, err)
			}
			outFile.Close()
		}
	}

	return combineDir, nil
}

// CountCombinedImagesByMaterialCodes 统计当前合成目录下每个素材编号对应的合成图片数量
func (s *ImageService) CountCombinedImagesByMaterialCodes(materialCodes []string) (map[string]int, error) {
	counts := make(map[string]int, len(materialCodes))
	for _, code := range materialCodes {
		materialCode := strings.TrimSpace(code)
		if materialCode == "" {
			continue
		}
		counts[materialCode] = 0
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("获取当前目录失败: %v", err)
	}

	combineDir := filepath.Join(currentDir, "合成")
	entries, err := os.ReadDir(combineDir)
	if err != nil {
		if os.IsNotExist(err) {
			return counts, nil
		}
		return nil, fmt.Errorf("读取合成目录失败: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		materialCode, ok := extractMaterialCodeFromCombinedFileName(entry.Name())
		if !ok {
			continue
		}

		if _, exists := counts[materialCode]; exists {
			counts[materialCode]++
		}
	}

	return counts, nil
}

func extractMaterialCodeFromCombinedFileName(fileName string) (string, bool) {
	if strings.ToLower(filepath.Ext(fileName)) != ".png" {
		return "", false
	}

	nameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	lastDashIndex := strings.LastIndex(nameWithoutExt, "-")
	if lastDashIndex <= 0 || lastDashIndex == len(nameWithoutExt)-1 {
		return "", false
	}

	if _, err := strconv.ParseInt(nameWithoutExt[lastDashIndex+1:], 10, 64); err != nil {
		return "", false
	}

	prefix := nameWithoutExt[:lastDashIndex]
	codeDashIndex := strings.LastIndex(prefix, "-")
	if codeDashIndex < 0 || codeDashIndex == len(prefix)-1 {
		return "", false
	}

	return strings.TrimSpace(prefix[codeDashIndex+1:]), true
}

// downloadImage 从 URL 下载图片
func downloadImage(url string) (image.Image, error) {
	// 创建带超时的 HTTP 客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("下载失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("解码图片失败: %v", err)
	}

	return img, nil
}

// CombineImages 使用一个或多个模板合成多张图片，并保存到当前目录下的"合成"目录
// 返回本次合成结果所在的目录路径
func (s *ImageService) CombineImages(templateIDs []string, sourcePaths []string) (string, error) {
	if len(templateIDs) == 0 {
		return "", fmt.Errorf("请选择至少一个模板")
	}
	if len(sourcePaths) == 0 {
		return "", fmt.Errorf("请选择至少一张图片")
	}

	// 合成目录: 当前目录/合成
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取当前目录失败: %v", err)
	}

	combineDir := filepath.Join(currentDir, "合成")
	if err := os.MkdirAll(combineDir, 0755); err != nil {
		return "", fmt.Errorf("创建合成目录失败: %v", err)
	}

	// 生成批次时间戳前缀，用于区分同一次合成的文件
	batchPrefix := time.Now().Format("20060102_150405")

	// 预加载选中的模板
	var selectedTemplates []*ImageTemplate
	for _, id := range templateIDs {
		template, err := findTemplateByID(id)
		if err != nil {
			return "", err
		}
		selectedTemplates = append(selectedTemplates, template)
	}

	for _, template := range selectedTemplates {
		// 加载当前模板底图
		templateImg, err := decodeImage(template.ImagePath)
		if err != nil {
			return "", fmt.Errorf("加载模板图片失败 (%s): %v", template.ImagePath, err)
		}

		for _, srcPath := range sourcePaths {
			// 加载源图片，支持 URL 和本地路径
			var srcImg image.Image
			if strings.HasPrefix(srcPath, "http://") || strings.HasPrefix(srcPath, "https://") {
				// URL，下载图片
				srcImg, err = downloadImage(srcPath)
				if err != nil {
					return "", fmt.Errorf("加载源图片失败 (%s): %v", srcPath, err)
				}
			} else {
				// 本地路径
				srcImg, err = decodeImage(srcPath)
				if err != nil {
					return "", fmt.Errorf("加载源图片失败 (%s): %v", srcPath, err)
				}
			}

			// 裁剪素材图片四周的空白区域
			trimmedSrcImg := trimImage(srcImg)

			outImg := compositeImage(
				templateImg,
				trimmedSrcImg,
				template.Width,
				template.Height,
				template.OffsetTop,
				template.OffsetRight,
				template.OffsetBottom,
				template.OffsetLeft,
			)

			// 提取文件名，去掉 URL 的查询参数，并为每个合成生成唯一文件名
			srcFileName := srcPath
			if strings.Contains(srcPath, "?") {
				srcFileName = srcPath[:strings.Index(srcPath, "?")]
			}
			baseName := filepath.Base(srcFileName)
			ext := strings.ToLower(filepath.Ext(baseName))
			nameWithoutExt := strings.TrimSuffix(baseName, ext)

			// 生成唯一文件名：模板名_时间戳_原文件名，确保不重复
			timestamp := time.Now().UnixNano()

			// 文件名增加模板名称前缀和时间戳，避免多模板时覆盖
			templateName := strings.TrimSpace(template.Name)
			if templateName == "" {
				templateName = template.ID
			}
			// 简单清理模板名中的路径分隔符
			templateName = strings.ReplaceAll(templateName, string(os.PathSeparator), "_")

			// 生成唯一文件名：批次前缀_模板名_时间戳_原文件名
			outName := fmt.Sprintf("%s_%s_%d_%s.png", batchPrefix, templateName, timestamp, nameWithoutExt)
			outPath := filepath.Join(combineDir, outName)

			outFile, err := os.Create(outPath)
			if err != nil {
				return "", fmt.Errorf("创建合成文件失败 (%s): %v", outPath, err)
			}

			if err := png.Encode(outFile, outImg); err != nil {
				outFile.Close()
				return "", fmt.Errorf("写入合成图片失败 (%s): %v", outPath, err)
			}
			outFile.Close()
		}
	}

	return combineDir, nil
}

// GetImageBase64 读取图片并转换为 base64，支持本地路径和 URL
func (s *ImageService) GetImageBase64(imagePath string) (string, error) {
	var data []byte
	var mimeType string

	// 判断是 URL 还是本地路径
	if strings.HasPrefix(imagePath, "http://") || strings.HasPrefix(imagePath, "https://") {
		// 从 URL 下载
		resp, err := http.Get(imagePath)
		if err != nil {
			return "", fmt.Errorf("下载图片失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("下载图片失败，状态码: %d", resp.StatusCode)
		}

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("读取图片数据失败: %v", err)
		}

		// 从 URL 或 Content-Type 推断 MIME 类型
		contentType := resp.Header.Get("Content-Type")
		if contentType != "" && strings.HasPrefix(contentType, "image/") {
			mimeType = contentType
		} else {
			// 从 URL 路径推断
			urlPath := imagePath
			if idx := strings.Index(imagePath, "?"); idx > 0 {
				urlPath = imagePath[:idx]
			}
			ext := strings.ToLower(filepath.Ext(urlPath))
			switch ext {
			case ".jpg", ".jpeg":
				mimeType = "image/jpeg"
			case ".png":
				mimeType = "image/png"
			case ".gif":
				mimeType = "image/gif"
			case ".bmp":
				mimeType = "image/bmp"
			case ".webp":
				mimeType = "image/webp"
			default:
				mimeType = "image/jpeg"
			}
		}
	} else {
		// 读取本地文件
		var err error
		data, err = os.ReadFile(imagePath)
		if err != nil {
			return "", fmt.Errorf("读取图片失败: %v", err)
		}

		// 检测图片类型
		ext := strings.ToLower(filepath.Ext(imagePath))
		switch ext {
		case ".jpg", ".jpeg":
			mimeType = "image/jpeg"
		case ".png":
			mimeType = "image/png"
		case ".gif":
			mimeType = "image/gif"
		case ".bmp":
			mimeType = "image/bmp"
		case ".webp":
			mimeType = "image/webp"
		default:
			mimeType = "image/jpeg"
		}
	}

	// 转换为 base64
	base64Str := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64Str), nil
}

// ServiceStartup 服务启动时调用
func (s *ImageService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	s.app = application.Get()
	return s.loadTemplates()
}
