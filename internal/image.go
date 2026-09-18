package internal

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// imageCache stores rendered ANSI strings keyed by path + dimensions + modtime
var (
	imageCache   = make(map[string]cachedImage)
	imageCacheMu sync.RWMutex
)

type cachedImage struct {
	modTime int64
	output  string
}

// ResolveImagePath searches for an image relative to baseDir, working directory, and standard subdirectories.
func ResolveImagePath(src string, baseDir string) (string, bool) {
	if src == "" {
		return "", false
	}

	var candidates []string

	if filepath.IsAbs(src) {
		candidates = append(candidates, src)
	} else {
		if baseDir != "" {
			candidates = append(candidates,
				filepath.Join(baseDir, src),
				filepath.Join(baseDir, "images", src),
				filepath.Join(baseDir, "assets", src),
				filepath.Join(baseDir, "TTP_Documentation", "development", "images", src),
			)
		}
		// Also try relative to current working directory
		candidates = append(candidates,
			src,
			filepath.Join("images", src),
			filepath.Join("assets", src),
		)
	}

	for _, path := range candidates {
		if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
			return path, true
		}
	}

	return "", false
}

// ImageInfo holds metadata about an image file.
type ImageInfo struct {
	Path   string
	Width  int
	Height int
	Format string
	Exists bool
}

// GetImageInfo queries basic dimensions and format for an image without full decoding.
func GetImageInfo(src, baseDir string) ImageInfo {
	fullPath, found := ResolveImagePath(src, baseDir)
	if !found {
		return ImageInfo{Path: src, Exists: false}
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return ImageInfo{Path: fullPath, Exists: true}
	}
	defer file.Close()

	cfg, format, err := image.DecodeConfig(file)
	if err != nil {
		return ImageInfo{Path: fullPath, Exists: true}
	}

	return ImageInfo{
		Path:   fullPath,
		Width:  cfg.Width,
		Height: cfg.Height,
		Format: strings.ToUpper(format),
		Exists: true,
	}
}

// RenderImage loads, scales, and renders an image as an ANSI truecolor half-block string.
func RenderImage(src string, baseDir string, maxWidth, maxHeight int) string {
	if maxWidth <= 0 {
		maxWidth = 40
	}
	if maxHeight <= 0 {
		maxHeight = 15
	}

	fullPath, found := ResolveImagePath(src, baseDir)
	if !found {
		return fmt.Sprintf("[ image not found: %s ]", src)
	}

	fi, err := os.Stat(fullPath)
	if err != nil {
		return fmt.Sprintf("[ image stat error: %v ]", err)
	}

	modTime := fi.ModTime().UnixNano()
	cacheKey := fmt.Sprintf("%s|%d|%d", fullPath, maxWidth, maxHeight)

	imageCacheMu.RLock()
	cached, ok := imageCache[cacheKey]
	imageCacheMu.RUnlock()

	if ok && cached.modTime == modTime {
		return cached.output
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return fmt.Sprintf("[ image open error: %v ]", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Sprintf("[ image decode error: %v ]", err)
	}

	output := renderImageToHalfBlocks(img, maxWidth, maxHeight)

	imageCacheMu.Lock()
	imageCache[cacheKey] = cachedImage{
		modTime: modTime,
		output:  output,
	}
	imageCacheMu.Unlock()

	return output
}

// renderImageToHalfBlocks converts an image.Image to ANSI truecolor half-block art.
// Each terminal cell holds 2 vertical pixels using the upper half block '▀' (U+2580).
// Cell width in columns = pixel width.
// Cell height in rows = pixel height / 2.
func renderImageToHalfBlocks(img image.Image, maxCols, maxRows int) string {
	bounds := img.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()

	if origW == 0 || origH == 0 {
		return "[ empty image ]"
	}

	// Calculate target dimensions in pixels
	// Terminal aspect ratio: characters are ~2x taller than wide.
	// Since 1 text cell has 2 vertical half-block pixels, 1 cell width = 1 cell height in pixels.
	maxPixelW := maxCols
	maxPixelH := maxRows * 2

	scaleW := float64(maxPixelW) / float64(origW)
	scaleH := float64(maxPixelH) / float64(origH)
	scale := scaleW
	if scaleH < scale {
		scale = scaleH
	}

	targetW := int(float64(origW) * scale)
	targetH := int(float64(origH) * scale)

	if targetW < 1 {
		targetW = 1
	}
	if targetH < 1 {
		targetH = 1
	}

	resized := resizeImage(img, targetW, targetH)

	var sb strings.Builder

	// Loop over rows in pairs (top pixel = y, bottom pixel = y+1)
	for y := 0; y < targetH; y += 2 {
		if y > 0 {
			sb.WriteString("\n")
		}

		var curFgR, curFgG, curFgB int = -1, -1, -1
		var curBgR, curBgG, curBgB int = -1, -1, -1

		for x := 0; x < targetW; x++ {
			topR, topG, topB, topA := colorToRGBA(resized.At(x, y))

			var botR, botG, botB, botA int
			if y+1 < targetH {
				botR, botG, botB, botA = colorToRGBA(resized.At(x, y+1))
			}

			topOpaque := topA > 32
			botOpaque := botA > 32

			if !topOpaque && !botOpaque {
				// Both transparent
				if curFgR != -1 || curBgR != -1 {
					sb.WriteString("\x1b[0m")
					curFgR, curFgG, curFgB = -1, -1, -1
					curBgR, curBgG, curBgB = -1, -1, -1
				}
				sb.WriteString(" ")
				continue
			}

			if topOpaque && !botOpaque {
				// Only top is opaque: use upper half block with transparent background
				if curBgR != -1 {
					sb.WriteString("\x1b[49m")
					curBgR, curBgG, curBgB = -1, -1, -1
				}
				if curFgR != topR || curFgG != topG || curFgB != topB {
					sb.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm", topR, topG, topB))
					curFgR, curFgG, curFgB = topR, topG, topB
				}
				sb.WriteString("▀")
				continue
			}

			if !topOpaque && botOpaque {
				// Only bottom is opaque: use lower half block '▄'
				if curBgR != -1 {
					sb.WriteString("\x1b[49m")
					curBgR, curBgG, curBgB = -1, -1, -1
				}
				if curFgR != botR || curFgG != botG || curFgB != botB {
					sb.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm", botR, botG, botB))
					curFgR, curFgG, curFgB = botR, botG, botB
				}
				sb.WriteString("▄")
				continue
			}

			// Both opaque: upper half block with top = Fg, bottom = Bg
			if curFgR != topR || curFgG != topG || curFgB != topB {
				sb.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm", topR, topG, topB))
				curFgR, curFgG, curFgB = topR, topG, topB
			}
			if curBgR != botR || curBgG != botG || curBgB != botB {
				sb.WriteString(fmt.Sprintf("\x1b[48;2;%d;%d;%dm", botR, botG, botB))
				curBgR, curBgG, curBgB = botR, botG, botB
			}
			sb.WriteString("▀")
		}

		// Reset at end of line
		sb.WriteString("\x1b[0m")
	}

	return sb.String()
}

// colorToRGBA extracts 8-bit RGBA components from a color.Color.
func colorToRGBA(c color.Color) (r, g, b, a int) {
	if c == nil {
		return 0, 0, 0, 0
	}
	r32, g32, b32, a32 := c.RGBA()
	if a32 == 0 {
		return 0, 0, 0, 0
	}
	// Undo color premultiplication
	r = int((r32 * 0xFF) / a32)
	g = int((g32 * 0xFF) / a32)
	b = int((b32 * 0xFF) / a32)
	a = int(a32 >> 8)
	return r, g, b, a
}

// resizeImage applies high-performance bilinear interpolation scaling.
func resizeImage(src image.Image, w, h int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	if w == 0 || h == 0 || srcW == 0 || srcH == 0 {
		return dst
	}

	xRatio := float64(srcW-1) / float64(w)
	yRatio := float64(srcH-1) / float64(h)

	for y := 0; y < h; y++ {
		srcY := float64(y) * yRatio
		yLow := int(srcY)
		yHigh := yLow + 1
		if yHigh >= srcH {
			yHigh = yLow
		}
		yWeight := srcY - float64(yLow)

		for x := 0; x < w; x++ {
			srcX := float64(x) * xRatio
			xLow := int(srcX)
			xHigh := xLow + 1
			if xHigh >= srcW {
				xHigh = xLow
			}
			xWeight := srcX - float64(xLow)

			c00 := src.At(srcBounds.Min.X+xLow, srcBounds.Min.Y+yLow)
			c10 := src.At(srcBounds.Min.X+xHigh, srcBounds.Min.Y+yLow)
			c01 := src.At(srcBounds.Min.X+xLow, srcBounds.Min.Y+yHigh)
			c11 := src.At(srcBounds.Min.X+xHigh, srcBounds.Min.Y+yHigh)

			r00, g00, b00, a00 := c00.RGBA()
			r10, g10, b10, a10 := c10.RGBA()
			r01, g01, b01, a01 := c01.RGBA()
			r11, g11, b11, a11 := c11.RGBA()

			// Bilinear interpolation
			topR := float64(r00)*(1-xWeight) + float64(r10)*xWeight
			botR := float64(r01)*(1-xWeight) + float64(r11)*xWeight
			r := uint8((topR*(1-yWeight) + botR*yWeight) / 257)

			topG := float64(g00)*(1-xWeight) + float64(g10)*xWeight
			botG := float64(g01)*(1-xWeight) + float64(g11)*xWeight
			g := uint8((topG*(1-yWeight) + botG*yWeight) / 257)

			topB := float64(b00)*(1-xWeight) + float64(b10)*xWeight
			botB := float64(b01)*(1-xWeight) + float64(b11)*xWeight
			b := uint8((topB*(1-yWeight) + botB*yWeight) / 257)

			topA := float64(a00)*(1-xWeight) + float64(a10)*xWeight
			botA := float64(a01)*(1-xWeight) + float64(a11)*xWeight
			a := uint8((topA*(1-yWeight) + botA*yWeight) / 257)

			dst.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}

	return dst
}
