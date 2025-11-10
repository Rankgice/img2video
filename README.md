# img2video

`img2video` 是一个命令行工具，可以通过重排源图像的像素来创建与目标图像相似的视觉效果。它可以将这个变换过程输出为 GIF 动画或单张静态图片。

这个项目的核心思想是：任何**两张尺寸完全相同**的图片，无论内容差异多大，它们的像素数据仅仅是排列方式不同。此工具通过计算最佳的像素移动路径，将源图片“变形”为目标图片。

## 功能

-   将源图片转换为目标图片的 GIF 动画。
-   生成一张由源图片像素重排而成的最终静态图 (PNG 格式)。
-   提供 `analyze` 命令来验证像素重排算法是否保持了像素数据的完整性。
-   支持两种不同的重排算法：`default` 和 `featured`。

## 使用方法

首先，请确保你已经编译了此项目。

```bash
go build .
```

### 命令

#### 1. 生成 GIF 动画

```bash
img2video gif <source_image> <target_image> <output.gif> [algorithm] [delay]
```

-   `<source_image>`: 源图片路径 (例如 `source.png`)。
-   `<target_image>`: 目标图片路径 (例如 `target.png`)。
-   `<output.gif>`: 输出的 GIF 文件名。
-   `[algorithm]` (可选): 使用的算法，可以是 `default` 或 `featured` (默认为 `default`)。
-   `[delay]` (可选): GIF 每帧之间的延迟，单位是百分之一秒 (默认为 1)。

#### 2. 生成静态图片

```bash
img2video image <source_image> <target_image> <output.png> [algorithm]
```

-   `<source_image>`: 源图片路径。
-   `<target_image>`: 目标图片路径。
-   `<output.png>`: 输出的 PNG 文件名。
-   `[algorithm]` (可选): 使用的算法，可以是 `default` 或 `featured` (默认为 `default`)。

#### 3. 分析算法

这个命令用于开发者验证像素重排算法是否正确。它会比较源图片和在内存中重排后的图片的灰度总和，如果两者一致，则证明算法没有丢失任何像素数据。（jpg格式是有损压缩，可能会导致检测出来不一致）

```bash
img2video analyze <source_image> <target_image> [algorithm]
```

-   `<source_image>`: 源图片路径。
-   `<target_image>`: 目标图片路径。
-   `[algorithm]` (可选): 要分析的算法，`default` 或 `featured`。

## 注意事项

1.  **图片尺寸**: 源图片和目标图片的尺寸（宽度和高度）必须完全相同。
2.  **文件格式**: 支持常见的图片格式，如 PNG, JPEG 等。
3.  **输出格式**:
    -   生成 GIF 时，由于 GIF 格式最多只支持 256 种颜色，程序会对颜色进行量化，这可能会导致最终动画的颜色与原图有轻微差异。
    -   生成静态图片时，推荐使用 PNG 格式输出，因为它是无损的，可以精确地保存重排后的像素颜色。
4.  **性能**: 处理大尺寸图片时，计算过程可能会消耗较多的时间和内存。
