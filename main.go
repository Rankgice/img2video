package main

import (
	"fmt"
	"image"
	_ "image/jpeg" // 导入 JPEG 解码器以支持解码
	_ "image/png"  // 导入 PNG 解码器以支持解码
	"log"
	"os"
	"strconv"
	"strings"
)

// readImage 从指定路径读取图片
func readImage(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file %s: %w", filePath, err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image file %s: %w", filePath, err)
	}
	return img, nil
}

func main() {
	// 更新使用说明
	if len(os.Args) < 5 {
		fmt.Printf("Usage: %s <source_image> <target_image> <output_type> <output_path> [delay]\n", os.Args[0])
		fmt.Println("  output_type: 'gif' or 'image'")
		fmt.Println("  output_path: path for the output file (e.g., animation.gif, result.png)")
		fmt.Println("  delay (for gif): optional, frame delay in 1/100s of a second (default: 1)")
		os.Exit(1)
	}

	sourceImagePath := os.Args[1]
	targetImagePath := os.Args[2]
	outputType := strings.ToLower(os.Args[3])
	outputPath := os.Args[4]
	frameDelay := 1 // 默认延迟

	if outputType == "gif" && len(os.Args) == 6 {
		delay, err := strconv.Atoi(os.Args[5])
		if err == nil && delay >= 1 {
			frameDelay = delay
		}
	}

	log.Printf("Reading source image: %s", sourceImagePath)
	sourceImg, err := readImage(sourceImagePath)
	if err != nil {
		log.Fatalf("Error reading source image: %v", err)
	}

	log.Printf("Reading target image: %s", targetImagePath)
	targetImg, err := readImage(targetImagePath)
	if err != nil {
		log.Fatalf("Error reading target image: %v", err)
	}

	if sourceImg.Bounds() != targetImg.Bounds() {
		log.Fatalf("Error: Source and target image dimensions must be the same.")
	}

	log.Println("Creating animation plan...")
	plan := CreateAnimationPlan(sourceImg, targetImg)

	switch outputType {
	case "gif":
		log.Println("Saving animation as GIF...")
		err := SaveGIF(plan, outputPath, frameDelay)
		if err != nil {
			log.Fatalf("Error saving GIF: %v", err)
		}
		log.Println("GIF animation created successfully!")
	case "image":
		log.Println("Saving final image...")
		err := SaveImage(plan, outputPath)
		if err != nil {
			log.Fatalf("Error saving image: %v", err)
		}
		log.Printf("Image saved successfully to: %s", outputPath)
	default:
		log.Fatalf("Unknown output type: %s. Please use 'gif' or 'image'.", outputType)
	}
}
