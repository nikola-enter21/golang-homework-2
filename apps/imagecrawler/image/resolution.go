package image

import (
	"encoding/xml"
	"fmt"
	"golang.org/x/image/webp"
	"golang.org/x/net/html/charset"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"unicode"
)

type SVG struct {
	XMLName xml.Name `xml:"svg"`
	Width   string   `xml:"width,attr"`
	Height  string   `xml:"height,attr"`
}

// Dimensions returns dimensions for png, jpg, jpeg
func Dimensions(body io.ReadCloser) (width, height string, err error) {
	defer body.Close()
	img, _, err := image.DecodeConfig(body)
	if err != nil {
		return "", "", err
	}

	return extractNumericPart(fmt.Sprint(img.Width)), extractNumericPart(fmt.Sprintln(img.Height)), nil
}

func SVGDimensions(body io.ReadCloser) (string, string, error) {
	defer body.Close()
	var svg SVG

	decoder := xml.NewDecoder(body)
	decoder.CharsetReader = charset.NewReaderLabel

	if err := decoder.Decode(&svg); err != nil {
		return "", "", err
	}

	return extractNumericPart(svg.Width), extractNumericPart(svg.Height), nil
}

func WebpDimensions(body io.ReadCloser) (width, height string, err error) {
	defer body.Close()

	decode, err := webp.Decode(body)
	if err != nil {
		return "", "", err
	}
	return extractNumericPart(fmt.Sprint(decode.Bounds().Dx())), extractNumericPart(fmt.Sprintln(decode.Bounds().Dy())), nil
}

func extractNumericPart(s string) string {
	numericPart := ""
	for _, char := range s {
		if unicode.IsDigit(char) {
			numericPart += string(char)
		}
	}
	return numericPart
}
