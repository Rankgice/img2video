package main

import (
	"fmt"
	"image"
	"image/color"
	"image/color/palette" // 导入标准颜色调色板
	"image/draw"          // 导入用于图像绘制的包
	"image/gif"           // 导入 GIF 编码器
	"image/jpeg"          // 导入 JPEG 解码器，用于读取 JPEG 图片
	"image/png"           // 导入 PNG 解码器，用于读取 PNG 图片
	"log"
	"os"
	"sort"
	"strconv" // 导入用于字符串到数字转换的包
)

// Pixel 结构体存储像素的灰度值、原始位置和原始颜色
type Pixel struct {
	GrayscaleValue float64
	OriginalX      int
	OriginalY      int
	Color          color.RGBA
}

// Pixels 是 Pixel 结构体的切片，用于实现 sort.Interface 接口
type Pixels []Pixel

// Len 是 sort.Interface 的方法，返回切片长度
func (p Pixels) Len() int { return len(p) }

// Less 是 sort.Interface 的方法，用于比较两个元素
func (p Pixels) Less(i, j int) bool { return p[i].GrayscaleValue < p[j].GrayscaleValue }

// Swap 是 sort.Interface 的方法，用于交换两个元素
func (p Pixels) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// abs 返回整数的绝对值
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// readImage 从指定路径读取图片
func readImage(filePath string) (image.Image, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open image file %s: %w", filePath, err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image file %s: %w", filePath, err)
	}
	return img, format, nil
}

// saveImage 将图片保存到指定路径
func saveImage(filePath string, img image.Image, format string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", filePath, err)
	}
	defer file.Close()

	switch format {
	case "png":
		return png.Encode(file, img)
	case "jpeg":
		// 可以设置 JPEG 质量，例如 &jpeg.Options{Quality: 90}
		return jpeg.Encode(file, img, nil)
	// 如果需要支持其他格式，可以在这里添加 case
	default:
		return png.Encode(file, img) // 默认保存为 PNG
	}
}

// imageToPixels 将 image.Image 转换为 Pixel 列表，并计算灰度值
func imageToPixels(img image.Image) ([]Pixel, error) {
	bounds := img.Bounds()
	pixels := make([]Pixel, 0, bounds.Dx()*bounds.Dy())

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := img.At(x, y)
			r, g, b, a := originalColor.RGBA() // RGBA 返回 uint32，范围 0-65535

			// 将 uint32 缩放到 uint8 范围 (0-255)
			c := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}

			// 计算灰度值 (Luma 亮度)
			// Y = 0.299*R + 0.587*G + 0.114*B (Rec. 601 Luma)
			grayscale := float64(c.R)*0.299 + float64(c.G)*0.587 + float64(c.B)*0.114

			pixels = append(pixels, Pixel{
				GrayscaleValue: grayscale,
				OriginalX:      x,
				OriginalY:      y,
				Color:          c,
			})
		}
	}
	return pixels, nil
}

// reorderImages 执行重排算法
func reorderImages(sourceImg, targetImg image.Image) (image.Image, error) {
	sourcePixels, err := imageToPixels(sourceImg)
	if err != nil {
		return nil, fmt.Errorf("failed to process source image pixels: %w", err)
	}
	targetPixels, err := imageToPixels(targetImg)
	if err != nil {
		return nil, fmt.Errorf("failed to process target image pixels: %w", err)
	}

	// 确保图片尺寸相同
	if len(sourcePixels) != len(targetPixels) {
		return nil, fmt.Errorf("source and target images must have the same number of pixels (dimensions)")
	}

	// 对像素列表进行排序
	sort.Sort(Pixels(sourcePixels))
	sort.Sort(Pixels(targetPixels))

	// 创建新的图像，尺寸与目标图相同
	bounds := targetImg.Bounds()
	reorderedImg := image.NewRGBA(bounds) // 使用 RGBA 格式创建新图像

	// 根据排序后的对应关系重排像素
	for i := 0; i < len(sourcePixels); i++ {
		// 目标图当前索引像素的原始位置，就是新图中要放置像素的位置
		targetPixelOriginalX := targetPixels[i].OriginalX
		targetPixelOriginalY := targetPixels[i].OriginalY

		// 原图当前索引像素的原始颜色，是要移动的颜色
		sourcePixelColor := sourcePixels[i].Color

		reorderedImg.Set(targetPixelOriginalX, targetPixelOriginalY, sourcePixelColor)
	}

	return reorderedImg, nil
}

func main() {
	// 检查命令行参数数量
	if len(os.Args) < 4 || len(os.Args) > 5 {
		fmt.Printf("Usage: %s <source_image_path> <target_image_path> <output_gif_path> [frame_delay_100ths_sec]\n", os.Args[0])
		fmt.Println("  frame_delay_100ths_sec: 可选参数。每帧之间的延迟，单位是 1/100 秒 (例如，10 表示 0.1 秒)。默认值是 1 (0.01 秒)。")
		os.Exit(1)
	}

	sourceImagePath := os.Args[1]
	targetImagePath := os.Args[2]
	outputGIFPath := os.Args[3]
	frameDelay := 1 // 默认延迟 0.01 秒 (1 个 1/100 秒)

	// 如果提供了可选的延迟参数，则解析它
	if len(os.Args) == 5 {
		delay, err := strconv.Atoi(os.Args[4])
		if err != nil {
			log.Printf("Warning: 无效的 frame_delay 参数 '%s'。将使用默认延迟 %d (0.01 秒)。错误信息: %v", os.Args[4], frameDelay, err)
		} else if delay < 1 {
			log.Printf("Warning: frame_delay 必须至少为 1。将使用默认延迟 %d (0.01 秒)。", frameDelay)
		} else {
			frameDelay = delay
		}
	}

	log.Printf("正在读取源图片: %s", sourceImagePath)
	sourceImg, _, err := readImage(sourceImagePath)
	if err != nil {
		log.Fatalf("读取源图片时出错: %v", err)
	}

	log.Printf("正在读取目标图片: %s", targetImagePath)
	targetImg, _, err := readImage(targetImagePath)
	if err != nil {
		log.Fatalf("读取目标图片时出错: %v", err)
	}

	// 确保图片尺寸相同
	if sourceImg.Bounds().Dx() != targetImg.Bounds().Dx() ||
		sourceImg.Bounds().Dy() != targetImg.Bounds().Dy() {
		log.Fatalf("错误: 源图片尺寸 (%dx%d) 与目标图片尺寸 (%dx%d) 不匹配。它们必须相同。",
			sourceImg.Bounds().Dx(), sourceImg.Bounds().Dy(), targetImg.Bounds().Dx(), targetImg.Bounds().Dy())
	}

	// 将图片转换为排序后的像素列表
	sourcePixels, err := imageToPixels(sourceImg)
	if err != nil {
		log.Fatalf("处理源图片像素时失败: %v", err)
	}
	targetPixels, err := imageToPixels(targetImg)
	if err != nil {
		log.Fatalf("处理目标图片像素时失败: %v", err)
	}

	// 对像素列表按灰度值进行排序
	sort.Sort(Pixels(sourcePixels))
	sort.Sort(Pixels(targetPixels))

	// 定义一个结构体来存储每个像素的动画详细信息
	type AnimationPixel struct {
		StartX  int
		StartY  int
		TargetX int
		TargetY int
		Color   color.RGBA
	}
	var animationPixels []AnimationPixel

	maxMoveSteps := 0 // 记录所有像素中最大的移动步数，这将是 GIF 的总帧数

	// 填充 animationPixels 并确定最大动画步数
	for i := 0; i < len(sourcePixels); i++ {
		ap := AnimationPixel{
			StartX:  sourcePixels[i].OriginalX,
			StartY:  sourcePixels[i].OriginalY,
			TargetX: targetPixels[i].OriginalX,
			TargetY: targetPixels[i].OriginalY,
			Color:   sourcePixels[i].Color,
		}
		dx := ap.TargetX - ap.StartX
		dy := ap.TargetY - ap.StartY
		currentPixelSteps := max(abs(dx), abs(dy)) // 计算当前像素所需的移动步数
		if currentPixelSteps > maxMoveSteps {
			maxMoveSteps = currentPixelSteps
		}
		animationPixels = append(animationPixels, ap)
	}

	// 准备 GIF 帧
	var gifFrames []*image.Paletted
	var gifDelays []int

	bounds := sourceImg.Bounds() // 所有帧都将使用相同的边界

	// 为 GIF 动画选择一个标准调色板。Plan9 提供 256 种颜色。
	// 对于最佳质量，可以从源图片和目标图片的所有颜色中生成一个自适应调色板。
	gifPalette := palette.Plan9

	log.Printf("正在生成 %d 帧动画，每帧延迟 %d (0.0%ds)... 对于大图片这可能需要一些时间。", maxMoveSteps+1, frameDelay, frameDelay)

	// 生成动画的每一帧
	for frameNum := 0; frameNum <= maxMoveSteps; frameNum++ {
		// 创建一个空白的 RGBA 图像用于当前帧
		currentFrameRGBA := image.NewRGBA(bounds)

		// 遍历所有像素，计算它们在当前帧的位置并绘制
		for _, ap := range animationPixels {
			currentX := ap.StartX
			currentY := ap.StartY

			dx := ap.TargetX - ap.StartX
			dy := ap.TargetY - ap.StartY

			// 独立计算此像素在 X 和 Y 方向已移动的步数
			movedX := min(frameNum, abs(dx))
			movedY := min(frameNum, abs(dy))

			// 更新当前 X 坐标
			if dx > 0 { // 如果目标 X 大于起点 X，则向右移动
				currentX = ap.StartX + movedX
			} else if dx < 0 { // 如果目标 X 小于起点 X，则向左移动
				currentX = ap.StartX - movedX
			}
			// 如果 dx == 0，currentX 保持不变

			// 更新当前 Y 坐标
			if dy > 0 { // 如果目标 Y 大于起点 Y，则向下移动
				currentY = ap.StartY + movedY
			} else if dy < 0 { // 如果目标 Y 小于起点 Y，则向上移动
				currentY = ap.StartY - movedY
			}
			// 如果 dy == 0，currentY 保持不变

			// 将像素绘制到当前帧图像上
			currentFrameRGBA.Set(currentX, currentY, ap.Color)
		}

		// 将 RGBA 帧转换为 Paletted 图像，用于 GIF 编码
		palettedFrame := image.NewPaletted(bounds, gifPalette)
		// draw.Draw 会执行颜色量化和抖动（如果需要）
		draw.Draw(palettedFrame, palettedFrame.Bounds(), currentFrameRGBA, image.Point{}, draw.Src)

		gifFrames = append(gifFrames, palettedFrame)
		gifDelays = append(gifDelays, frameDelay)

		if frameNum%(maxMoveSteps/10+1) == 0 { // 每生成一定比例的帧数就记录一次进度
			log.Printf("已生成帧 %d/%d", frameNum, maxMoveSteps)
		}
	}

	// 将 GIF 写入文件
	outputFile, err := os.Create(outputGIFPath)
	if err != nil {
		log.Fatalf("创建输出 GIF 文件 %s 时出错: %v", outputGIFPath, err)
	}
	defer outputFile.Close()

	g := &gif.GIF{
		Image:     gifFrames,
		Delay:     gifDelays,
		LoopCount: 0, // 0 表示无限循环播放
	}

	log.Printf("正在将 GIF 动画编码到 %s...", outputGIFPath)
	err = gif.EncodeAll(outputFile, g)
	if err != nil {
		log.Fatalf("编码 GIF 时出错: %v", err)
	}

	log.Println("GIF 动画已成功创建！")
}
