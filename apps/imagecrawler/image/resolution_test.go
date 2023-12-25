package image

import (
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestGIFDimensions(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	svgPath := filepath.Join(currentDir, "horse.gif")
	file, err := os.Open(svgPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	width, height, err := Dimensions(file)
	if err != nil {
		t.Fatal(err)
	}

	expectedWidth := "307"
	expectedHeight := "230"
	require.Equal(t, expectedWidth, width)
	require.Equal(t, expectedHeight, height)
}

func TestSVGDimensions(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	svgPath := filepath.Join(currentDir, "apple-logo.svg")
	file, err := os.Open(svgPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	width, height, err := SVGDimensions(file)
	if err != nil {
		t.Fatal(err)
	}

	expectedWidth := "800"
	expectedHeight := "800"
	require.Equal(t, expectedWidth, width)
	require.Equal(t, expectedHeight, height)
}

func TestWEBPDimensions(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	svgPath := filepath.Join(currentDir, "forest.webp")
	file, err := os.Open(svgPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	width, height, err := WebpDimensions(file)
	if err != nil {
		t.Fatal(err)
	}

	expectedWidth := "320"
	expectedHeight := "214"
	require.Equal(t, expectedWidth, width)
	require.Equal(t, expectedHeight, height)
}
