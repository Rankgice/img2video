package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif" // 导入 GIF 编码/解码器 (如果需要支持 GIF)
	"image/jpeg"  // 导入 JPEG 编码/解码器 (如果需要支持 JPEG)
	"image/png"   // 导入 PNG 编码/解码器
	"log"
	"os"
	"sort"
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
			r, g, b, a := originalColor.RGBA() // RGBA returns uint32, 0-65535

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
	if len(os.Args) != 4 {
		fmt.Printf("Usage: %s <source_image_path> <target_image_path> <output_image_path>\n", os.Args[0])
		os.Exit(1)
	}

	sourceImagePath := os.Args[1]
	targetImagePath := os.Args[2]
	outputImagePath := os.Args[3]

	log.Printf("Reading source image: %s", sourceImagePath)
	sourceImg, sourceFormat, err := readImage(sourceImagePath)
	if err != nil {
		log.Fatalf("Error reading source image: %v", err)
	}
	log.Printf("Source image format: %s", sourceFormat)

	log.Printf("Reading target image: %s", targetImagePath)
	targetImg, targetFormat, err := readImage(targetImagePath)
	if err != nil {
		log.Fatalf("Error reading target image: %v", err)
	}
	log.Printf("Target image format: %s", targetFormat)

	// 再次检查尺寸，虽然在 reorderImages 中有，但在主函数中明确提示更友好
	if sourceImg.Bounds().Dx() != targetImg.Bounds().Dx() ||
		sourceImg.Bounds().Dy() != targetImg.Bounds().Dy() {
		log.Fatalf("Error: Source image dimensions (%dx%d) do not match target image dimensions (%dx%d). They must be identical.",
			sourceImg.Bounds().Dx(), sourceImg.Bounds().Dy(), targetImg.Bounds().Dx(), targetImg.Bounds().Dy())
	}

	log.Println("Starting image reordering...")
	reorderedImg, err := reorderImages(sourceImg, targetImg)
	if err != nil {
		log.Fatalf("Error reordering images: %v", err)
	}
	log.Println("Image reordering completed.")

	// 保存时使用源图片的格式，如果不支持则默认PNG
	log.Printf("Saving reordered image to: %s (format: %s)", outputImagePath, sourceFormat)
	err = saveImage(outputImagePath, reorderedImg, sourceFormat)
	if err != nil {
		log.Fatalf("Error saving reordered image: %v", err)
	}

	log.Println("Done!")
}
