package main

import (
	"image"
	"image/color"
	"image/draw"
	"sort"
)

// Pixel 结构体存储像素的灰度值、原始位置和原始颜色
type Pixel struct {
	GrayscaleValue float64
	OriginalX      int
	OriginalY      int
	Color          color.RGBA
}

// Pixels 是 Pixel 结构体的切片，用于实现 sort.Interface 接口（复杂排序）
type Pixels []Pixel

func (p Pixels) Len() int { return len(p) }
func (p Pixels) Less(i, j int) bool { // 首先按灰度值排序
	if p[i].GrayscaleValue != p[j].GrayscaleValue {
		return p[i].GrayscaleValue < p[j].GrayscaleValue
	}
	// 如果灰度值相同，按绿色分量排序
	if p[i].Color.G != p[j].Color.G {
		return p[i].Color.G < p[j].Color.G
	}
	// 如果绿色分量也相同，按红色分量排序
	return p[i].Color.R < p[j].Color.R
}
func (p Pixels) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// PixelFeatured 结构体用于特征排序，增加了区间深度字段
type PixelFeatured struct {
	Pixel
	IntervalDepth float64
}

// PixelsFeatured 是 PixelFeatured 的切片，用于实现特征排序
type PixelsFeatured []PixelFeatured

func (p PixelsFeatured) Len() int      { return len(p) }
func (p PixelsFeatured) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p PixelsFeatured) Less(i, j int) bool {
	if p[i].GrayscaleValue != p[j].GrayscaleValue {
		return p[i].GrayscaleValue < p[j].GrayscaleValue
	}
	return p[i].IntervalDepth < p[j].IntervalDepth
}

// AnimationPixel 存储每个像素的动画详细信息
type AnimationPixel struct {
	StartX  int
	StartY  int
	TargetX int
	TargetY int
	Color   color.RGBA
}

// AnimationPlan 存储生成动画所需的所有计算数据
type AnimationPlan struct {
	Pixels []AnimationPixel
	Frames int
	Bounds image.Rectangle
}

// imageToPixels 将 image.Image 转换为 Pixel 列表，并计算灰度值
func imageToPixels(img image.Image) []Pixel {
	bounds := img.Bounds()
	pixels := make([]Pixel, 0, bounds.Dx()*bounds.Dy())
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := img.At(x, y)
			r, g, b, a := originalColor.RGBA()
			c := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
			grayscale := float64(c.R)*0.299 + float64(c.G)*0.587 + float64(c.B)*0.114
			pixels = append(pixels, Pixel{
				GrayscaleValue: grayscale,
				OriginalX:      x,
				OriginalY:      y,
				Color:          c,
			})
		}
	}
	return pixels
}

// calculatePlan 是一个辅助函数，用于根据排序后的像素列表计算动画计划
func calculatePlan(sourcePixels []Pixel, targetPixels []PixelFeatured, bounds image.Rectangle) *AnimationPlan {
	var animationPixels []AnimationPixel
	maxMoveSteps := 0
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
		currentPixelSteps := max(abs(dx), abs(dy))
		if currentPixelSteps > maxMoveSteps {
			maxMoveSteps = currentPixelSteps
		}
		animationPixels = append(animationPixels, ap)
	}
	return &AnimationPlan{
		Pixels: animationPixels,
		Frames: maxMoveSteps + 1,
		Bounds: bounds,
	}
}

// CreateAnimationPlan 计算源图像到目标图像的像素移动路径（默认复杂排序）
func CreateAnimationPlan(sourceImg, targetImg image.Image) *AnimationPlan {
	sourcePixels := imageToPixels(sourceImg)
	targetPixels := imageToPixels(targetImg)
	sort.Sort(Pixels(sourcePixels))
	sort.Sort(Pixels(targetPixels))

	// 将 targetPixels 转换为 PixelFeatured 以匹配 calculatePlan 的签名
	targetPixelsFeatured := make([]PixelFeatured, len(targetPixels))
	for i, p := range targetPixels {
		targetPixelsFeatured[i] = PixelFeatured{Pixel: p}
	}
	return calculatePlan(sourcePixels, targetPixelsFeatured, sourceImg.Bounds())
}

// CreateAnimationPlanFeatured 使用特征排序计算动画计划
func CreateAnimationPlanFeatured(sourceImg, targetImg image.Image) *AnimationPlan {
	sourcePixels := imageToPixels(sourceImg)
	targetPixelsRaw := imageToPixels(targetImg)

	// 1. 对源图使用默认复杂排序
	sort.Sort(Pixels(sourcePixels))

	// 2. 对目标图使用特征排序
	// 2a. 预计算灰度网格以便快速查找
	bounds := targetImg.Bounds()
	// 先将目标图转为灰度图
	grayImg := image.NewGray(bounds)
	draw.Draw(grayImg, bounds, targetImg, bounds.Min, draw.Src)

	grayGrid := make([][]float64, bounds.Max.Y)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		grayGrid[y] = make([]float64, bounds.Max.X)
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// 从灰度图中安全地读取灰度值
			grayGrid[y][x] = float64(grayImg.GrayAt(x, y).Y)
		}
	}

	// 2b. 计算每个目标像素的区间深度
	targetPixelsFeatured := make([]PixelFeatured, len(targetPixelsRaw))
	for i, p := range targetPixelsRaw {
		depth := calculateIntervalDepth(p.OriginalX, p.OriginalY, grayGrid, bounds)
		targetPixelsFeatured[i] = PixelFeatured{Pixel: p, IntervalDepth: depth}
	}

	// 2c. 对目标像素进行特征排序
	sort.Sort(PixelsFeatured(targetPixelsFeatured))

	return calculatePlan(sourcePixels, targetPixelsFeatured, sourceImg.Bounds())
}

// calculateIntervalDepth 计算给定坐标的像素的区间深度
func calculateIntervalDepth(x, y int, grayGrid [][]float64, bounds image.Rectangle) float64 {
	avg3x3 := calculateAverageGray(x, y, 1, grayGrid, bounds) // 3x3 区域半径为 1
	avg5x5 := calculateAverageGray(x, y, 2, grayGrid, bounds) // 5x5 区域半径为 2
	return avg5x5*0.25 + avg3x3*0.75
}

// CalculateGrayscaleSum 计算并返回图像所有像素的灰度值总和
func CalculateGrayscaleSum(img image.Image) float64 {
	var sum float64
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// 使用与 imageToPixels 中相同的亮度计算公式
			grayscale := float64(uint8(r>>8))*0.299 + float64(uint8(g>>8))*0.587 + float64(uint8(b>>8))*0.114
			sum += grayscale
		}
	}
	return sum
}

// calculateAverageGray 计算以 (cx, cy) 为中心，半径为 radius 的区域的平均灰度值
func calculateAverageGray(cx, cy, radius int, grayGrid [][]float64, bounds image.Rectangle) float64 {
	var sum float64
	var count int
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			// 确保坐标在图像边界内
			if x >= bounds.Min.X && x < bounds.Max.X && y >= bounds.Min.Y && y < bounds.Max.Y {
				sum += grayGrid[y][x]
				count++
			}
		}
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}
