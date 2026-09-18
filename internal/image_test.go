package internal

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveImagePath(t *testing.T) {
	// Missing file should return false
	_, found := ResolveImagePath("non_existent_image_12345.png", "")
	if found {
		t.Errorf("expected missing image not to be found")
	}

	// Create a temporary directory with a test image
	tmpDir := t.TempDir()
	testImgPath := filepath.Join(tmpDir, "test.png")

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	f, err := os.Create(testImgPath)
	if err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatalf("failed to encode test image: %v", err)
	}
	f.Close()

	// Resolving relative to baseDir
	resolved, found := ResolveImagePath("test.png", tmpDir)
	if !found {
		t.Fatalf("expected test.png to be found in %s", tmpDir)
	}
	if resolved != testImgPath {
		t.Errorf("expected path %q, got %q", testImgPath, resolved)
	}
}

func TestRenderImageMissing(t *testing.T) {
	out := RenderImage("totally_missing.png", "", 40, 15)
	if !strings.Contains(out, "image not found") {
		t.Errorf("expected placeholder with 'image not found', got %q", out)
	}
}

func TestRenderImageOutput(t *testing.T) {
	tmpDir := t.TempDir()
	testImgPath := filepath.Join(tmpDir, "red_square.png")

	// Create a 4x4 red image
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	f, err := os.Create(testImgPath)
	if err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatalf("failed to encode test image: %v", err)
	}
	f.Close()

	out := RenderImage("red_square.png", tmpDir, 10, 10)
	if !strings.Contains(out, "▀") {
		t.Errorf("expected rendered half-block output to contain '▀', got %q", out)
	}
	if !strings.Contains(out, "\x1b[38;2;255;0;0m") {
		t.Errorf("expected ANSI truecolor escape for red, got %q", out)
	}
}

func TestRenderRealImages(t *testing.T) {
	// Test rendering demo.png in root
	out := RenderImage("demo.png", "..", 40, 12)
	if strings.Contains(out, "image not found") {
		t.Fatalf("demo.png should be found, got: %s", out)
	}
	if !strings.Contains(out, "▀") {
		t.Errorf("expected rendered demo.png to contain half-blocks, got: %s", out)
	}

	// Test rendering 1.png in TTP_Documentation/development/images
	out2 := RenderImage("1.png", "..", 40, 12)
	if strings.Contains(out2, "image not found") {
		t.Fatalf("1.png should be found via fallback path, got: %s", out2)
	}
	if !strings.Contains(out2, "▀") {
		t.Errorf("expected rendered 1.png to contain half-blocks, got: %s", out2)
	}
}

func TestGetImageInfo(t *testing.T) {
	info := GetImageInfo("demo.png", "..")
	if !info.Exists {
		t.Fatalf("expected demo.png to exist")
	}
	if info.Format != "PNG" {
		t.Errorf("expected format PNG, got %q", info.Format)
	}
	if info.Width <= 0 || info.Height <= 0 {
		t.Errorf("expected positive dimensions, got %dx%d", info.Width, info.Height)
	}
}
