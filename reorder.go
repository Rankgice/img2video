package main

import (
	"image"
	"image/color"
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

// abs 返回整数的绝对值
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// CreateAnimationPlan 计算源图像到目标图像的像素移动路径
func CreateAnimationPlan(sourceImg, targetImg image.Image) *AnimationPlan {
	sourcePixels := imageToPixels(sourceImg)
	targetPixels := imageToPixels(targetImg)

	sort.Sort(Pixels(sourcePixels))
	sort.Sort(Pixels(targetPixels))

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
		Frames: maxMoveSteps + 1, // 总帧数是最大步数+1（包含起始帧）
		Bounds: sourceImg.Bounds(),
	}
}
