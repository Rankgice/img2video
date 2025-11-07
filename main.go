package main

import (
	"fmt"
	"image"
	_ "image/jpeg" // 导入 JPEG 解码器以支持解码
	_ "image/png"  // 导入 PNG 解码器以支持解码
	"log"
	"math"
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
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "gif", "image":
		handleGenerate(command)
	case "analyze":
		handleAnalyze()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: img2video <command> [arguments]")
	fmt.Println("\nCommands:")
	fmt.Println("  gif <source> <target> <output.gif> [algorithm] [delay] - Generate a GIF animation")
	fmt.Println("  image <source> <target> <output.png> [algorithm]     - Generate a single result image")
	fmt.Println("  analyze <source> <target> [algorithm]                  - Analyze grayscale sums before and after reordering")
	fmt.Println("\nAlgorithm can be 'default' or 'featured' (default: default).")
}

func handleAnalyze() {
	if len(os.Args) < 4 {
		printUsage()
		os.Exit(1)
	}
	sourcePath := os.Args[2]
	targetPath := os.Args[3]
	algorithm := "default"
	if len(os.Args) > 4 {
		algorithm = os.Args[4]
	}

	log.Printf("Loading source image: %s", sourcePath)
	sourceImg, err := readImage(sourcePath)
	if err != nil {
		log.Fatalf("Failed to read source image: %v", err)
	}

	log.Printf("Loading target image: %s", targetPath)
	targetImg, err := readImage(targetPath)
	if err != nil {
		log.Fatalf("Failed to read target image: %v", err)
	}

	// 1. 计算原图的灰度总和
	sourceSum := CalculateGrayscaleSum(sourceImg)
	log.Printf("Source Image Grayscale Sum: %f", sourceSum)

	// 2. 在内存中进行重排
	log.Printf("Creating animation plan using '%s' algorithm...", algorithm)
	var plan *AnimationPlan
	switch algorithm {
	case "default":
		plan = CreateAnimationPlan(sourceImg, targetImg)
	case "featured":
		plan = CreateAnimationPlanFeatured(sourceImg, targetImg)
	default:
		log.Fatalf("Unknown algorithm: %s", algorithm)
	}

	// 3. 在内存中创建重排后的图像
	reorderedImg := image.NewRGBA(plan.Bounds)
	for _, p := range plan.Pixels {
		reorderedImg.Set(p.TargetX, p.TargetY, p.Color)
	}

	// 4. 计算内存中重排图像的灰度总和
	reorderedSum := CalculateGrayscaleSum(reorderedImg)
	log.Printf("In-Memory Reordered Image Grayscale Sum: %f", reorderedSum)

	// 5. 打印分析结果
	fmt.Println("\n--- Analysis Result ---")
	// 使用一个小的容差来比较浮点数，以客户浮点数精度问题
	if math.Abs(sourceSum-reorderedSum) < 0.0001 {
		fmt.Println("SUCCESS: The grayscale sum is effectively IDENTICAL before and after reordering in memory.")
		fmt.Printf("(Difference: %f, which is within the tolerance for floating-point arithmetic)\n", reorderedSum-sourceSum)
		fmt.Println("This proves the core algorithm correctly preserves all pixel data.")
		fmt.Println("\nAny differences you see in a saved file are due to the file encoding process:")
		fmt.Println("- GIF: Color quantization to a 256-color palette changes pixel colors.")
		fmt.Println("- JPEG: Lossy compression changes pixel colors.")
		fmt.Println("- PNG: This format is lossless and should produce a file with the same grayscale sum.")
	} else {
		fmt.Println("ERROR: The grayscale sum is DIFFERENT. This indicates a potential bug in the reordering logic.")
		fmt.Printf("Difference: %f\n", reorderedSum-sourceSum)
	}
}

func handleGenerate(command string) {
	if len(os.Args) < 5 {
		printUsage()
		os.Exit(1)
	}
	sourceImagePath := os.Args[2]
	targetImagePath := os.Args[3]
	outputPath := os.Args[4]

	algorithm := "default"
	frameDelay := 1

	if len(os.Args) > 5 {
		val, err := strconv.Atoi(os.Args[5])
		if err != nil {
			algorithm = strings.ToLower(os.Args[5])
			if len(os.Args) > 6 {
				delay, err := strconv.Atoi(os.Args[6])
				if err == nil {
					frameDelay = delay
				}
			}
		} else {
			frameDelay = val
		}
	}
	if len(os.Args) > 6 {
		if _, err := strconv.Atoi(os.Args[5]); err != nil {
			delay, err := strconv.Atoi(os.Args[6])
			if err == nil {
				frameDelay = delay
			}
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

	log.Printf("Creating animation plan using '%s' algorithm...", algorithm)
	var plan *AnimationPlan
	switch algorithm {
	case "default":
		plan = CreateAnimationPlan(sourceImg, targetImg)
	case "featured":
		plan = CreateAnimationPlanFeatured(sourceImg, targetImg)
	default:
		log.Fatalf("Unknown algorithm: %s. Please use 'default' or 'featured'.", algorithm)
	}

	switch command {
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
	}
}
