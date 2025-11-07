package main

import (
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// SaveGIF 根据 AnimationPlan 生成并保存 GIF 动画（随机步长）
func SaveGIF(plan *AnimationPlan, outputPath string, delay int) error {
	rand.Seed(time.Now().UnixNano())

	var gifFrames []*image.Paletted
	var gifDelays []int
	gifPalette := palette.Plan9

	// 存储每个像素的当前位置
	type currentPixelState struct {
		X, Y int
	}
	pixelStates := make([]currentPixelState, len(plan.Pixels))
	for i, p := range plan.Pixels {
		pixelStates[i] = currentPixelState{X: p.StartX, Y: p.StartY}
	}

	log.Println("正在生成随机步长动画...")

	// 首先，将原图作为第一帧
	firstFrame := image.NewRGBA(plan.Bounds)
	for _, p := range plan.Pixels {
		firstFrame.Set(p.StartX, p.StartY, p.Color)
	}
	palettedFirstFrame := image.NewPaletted(plan.Bounds, gifPalette)
	draw.Draw(palettedFirstFrame, palettedFirstFrame.Bounds(), firstFrame, image.Point{}, draw.Src)
	gifFrames = append(gifFrames, palettedFirstFrame)
	gifDelays = append(gifDelays, delay) // 可以为第一帧设置不同的延迟，这里使用相同延迟

	frameCount := 1 // 从第1帧开始计数（因为第0帧已经是原图）
	for {
		frameCount++
		allArrived := true
		currentFrameRGBA := image.NewRGBA(plan.Bounds)

		for i, ap := range plan.Pixels {
			state := &pixelStates[i]

			// 如果还没到达，就移动它
			if state.X != ap.TargetX || state.Y != ap.TargetY {
				allArrived = false

				// 计算到目标的距离
				dx := ap.TargetX - state.X
				dy := ap.TargetY - state.Y

				// 根据图片尺寸计算缩放因子
				scaleX := float64(plan.Bounds.Dx()) / 150.0
				scaleY := float64(plan.Bounds.Dy()) / 150.0

				// 获取基础随机步长 (1-3)
				baseStepX := rand.Intn(3) + 1
				baseStepY := rand.Intn(3) + 1

				// 计算最终步长，并确保至少为 1
				stepX := max(max(1, int(scaleX)), int(math.Round(float64(baseStepX)*scaleX)))
				stepY := max(max(1, int(scaleY), int(math.Round(float64(baseStepY)*scaleY))))

				// 移动 X 轴
				if abs(dx) <= stepX {
					state.X = ap.TargetX
				} else if dx > 0 {
					state.X += stepX
				} else {
					state.X -= stepX
				}

				// 移动 Y 轴
				if abs(dy) <= stepY {
					state.Y = ap.TargetY
				} else if dy > 0 {
					state.Y += stepY
				} else {
					state.Y -= stepY
				}
			}
			currentFrameRGBA.Set(state.X, state.Y, ap.Color)
		}

		// 将帧添加到 GIF
		palettedFrame := image.NewPaletted(plan.Bounds, gifPalette)
		draw.Draw(palettedFrame, palettedFrame.Bounds(), currentFrameRGBA, image.Point{}, draw.Src)
		gifFrames = append(gifFrames, palettedFrame)
		gifDelays = append(gifDelays, delay)

		if frameCount%20 == 0 {
			log.Printf("已生成 %d 帧...", frameCount)
		}

		if allArrived {
			log.Printf("所有像素已到达，总共生成 %d 帧。", frameCount)
			break
		}
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出 GIF 文件 %s 时出错: %w", outputPath, err)
	}
	defer outputFile.Close()

	g := &gif.GIF{
		Image:     gifFrames,
		Delay:     gifDelays,
		LoopCount: 0, // 0 表示无限循环
	}

	log.Printf("正在将 GIF 动画编码到 %s...", outputPath)
	return gif.EncodeAll(outputFile, g)
}

// SaveImage 根据 AnimationPlan 生成并保存最终的重排图像
func SaveImage(plan *AnimationPlan, outputPath string) error {
	log.Printf("正在生成最终的重排图像...")

	finalImage := image.NewRGBA(plan.Bounds)

	for _, ap := range plan.Pixels {
		// 在最后一帧，所有像素都应在其目标位置
		finalImage.Set(ap.TargetX, ap.TargetY, ap.Color)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件 %s 时出错: %w", outputPath, err)
	}
	defer file.Close()

	log.Printf("正在将图像编码到 %s...", outputPath)
	// 根据文件扩展名选择编码器，默认为 PNG
	ext := filepath.Ext(outputPath)
	if ext == ".jpg" || ext == ".jpeg" {
		// 可以为 JPEG 设置质量选项
		return jpeg.Encode(file, finalImage, nil)
	}
	return png.Encode(file, finalImage)
}
