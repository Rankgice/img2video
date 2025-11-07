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
	"os"
	"path/filepath"
)

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SaveGIF 根据 AnimationPlan 生成并保存 GIF 动画
func SaveGIF(plan *AnimationPlan, outputPath string, delay int) error {
	var gifFrames []*image.Paletted
	var gifDelays []int

	gifPalette := palette.Plan9
	log.Printf("正在生成 %d 帧动画...", plan.Frames)

	for frameNum := 0; frameNum < plan.Frames; frameNum++ {
		currentFrameRGBA := image.NewRGBA(plan.Bounds)

		for _, ap := range plan.Pixels {
			currentX, currentY := calculatePixelPosition(ap, frameNum)
			currentFrameRGBA.Set(currentX, currentY, ap.Color)
		}

		palettedFrame := image.NewPaletted(plan.Bounds, gifPalette)
		draw.Draw(palettedFrame, palettedFrame.Bounds(), currentFrameRGBA, image.Point{}, draw.Src)

		gifFrames = append(gifFrames, palettedFrame)
		gifDelays = append(gifDelays, delay)

		if frameNum%(plan.Frames/10+1) == 0 {
			log.Printf("已生成帧 %d/%d", frameNum, plan.Frames-1)
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

// calculatePixelPosition 计算像素在特定帧的当前位置
func calculatePixelPosition(ap AnimationPixel, frameNum int) (int, int) {
	currentX := ap.StartX
	currentY := ap.StartY

	dx := ap.TargetX - ap.StartX
	dy := ap.TargetY - ap.StartY

	movedX := min(frameNum, abs(dx))
	movedY := min(frameNum, abs(dy))

	if dx > 0 {
		currentX = ap.StartX + movedX
	} else if dx < 0 {
		currentX = ap.StartX - movedX
	}

	if dy > 0 {
		currentY = ap.StartY + movedY
	} else if dy < 0 {
		currentY = ap.StartY - movedY
	}

	return currentX, currentY
}
