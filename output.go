package main

import (
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
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

// SaveFrames 将动画的每一帧保存为单独的 PNG 图片
func SaveFrames(plan *AnimationPlan, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建帧输出目录 %s 时出错: %w", outputDir, err)
	}

	log.Printf("正在将 %d 帧保存到目录 %s...", plan.Frames, outputDir)

	for frameNum := 0; frameNum < plan.Frames; frameNum++ {
		currentFrameRGBA := image.NewRGBA(plan.Bounds)

		for _, ap := range plan.Pixels {
			currentX, currentY := calculatePixelPosition(ap, frameNum)
			currentFrameRGBA.Set(currentX, currentY, ap.Color)
		}

		framePath := filepath.Join(outputDir, fmt.Sprintf("frame_%05d.png", frameNum))
		file, err := os.Create(framePath)
		if err != nil {
			return fmt.Errorf("创建帧文件 %s 时出错: %w", framePath, err)
		}

		err = png.Encode(file, currentFrameRGBA)
		file.Close() // 确保文件在检查错误前关闭
		if err != nil {
			return fmt.Errorf("编码帧 %s 时出错: %w", framePath, err)
		}

		if frameNum%(plan.Frames/10+1) == 0 {
			log.Printf("已保存帧 %d/%d", frameNum, plan.Frames-1)
		}
	}
	return nil
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
