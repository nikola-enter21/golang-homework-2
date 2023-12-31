package extracter

import (
	"awesomeProject1/apps/imagecrawler/image"
	"awesomeProject1/db"
	"awesomeProject1/db/model"
	"awesomeProject1/logger"
	"context"
	"errors"
	"github.com/google/uuid"
	urllib "net/url"
	"sync"

	"golang.org/x/net/html"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	log = logger.New()
)

func ExtractLinks(ctx context.Context, bodyBytes []byte, fullURL string) ([]string, error) {
	url := mustParseURL(fullURL)
	links := make([]string, 0)

	tokenizer := html.NewTokenizer(strings.NewReader(string(bodyBytes)))
	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if tokenizer.Err() != nil && tokenizer.Err() != io.EOF {
				return nil, tokenizer.Err()
			}
			return links, nil // Finished parsing
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()
			if token.Data == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						hrefLink := attr.Val
						if strings.HasPrefix(hrefLink, "http") || strings.HasPrefix(hrefLink, "https") {
							links = append(links, hrefLink)
						} else {
							absoluteURL := url.ResolveReference(mustParseURL(hrefLink)).String()
							links = append(links, absoluteURL)
						}
					}
				}
			}
		}
	}
}

func DownloadImages(
	ctx context.Context,
	imagesFolderName string,
	db db.DB,
	bodyBytes []byte,
	fullURL string,
	downloaded *sync.Map,
) error {
	tokenizer := html.NewTokenizer(strings.NewReader(string(bodyBytes)))
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if tokenizer.Err() != nil && tokenizer.Err() != io.EOF {
				return tokenizer.Err()
			}
			return nil // Finished parsing
		case html.SelfClosingTagToken, html.StartTagToken:
			token := tokenizer.Token()
			if token.Data == "img" {
				img, err := extractImageFromToken(token, fullURL)
				if err != nil {
					log.Infof("Extracting image error: %s", err)
					continue
				}
				if len(img.SourceURL) == 0 {
					log.Infof("Image does not have src attribute: %s", token.String())
				}
				_, alreadyDownloaded := downloaded.LoadOrStore(img.SourceURL, true)
				if alreadyDownloaded {
					continue
				}
				err = saveImage(ctx, imagesFolderName, db, img)
				if err != nil {
					log.Infof("Saving image error: %s", err)
					continue
				}
			}
		}
	}
}

// mustParseURL parses a URL and panics on error.
func mustParseURL(rawURL string) *urllib.URL {
	u, err := urllib.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}

func extractImageFromToken(token html.Token, fullURL string) (*model.Image, error) {
	img := &model.Image{}
	for _, attr := range token.Attr {
		switch attr.Key {
		case "src":
			if strings.HasPrefix(attr.Val, "data:") {
				return nil, errors.New("crawler does not support base64 images")
			}

			imgURL := mustParseURL(attr.Val)
			imgURL = resolveURL(fullURL, imgURL)
			response, err := http.Get(imgURL.String())
			if err != nil {
				return nil, err
			}

			img.SourceURL = imgURL.String()
			img.Filename = uuid.NewString()
			format, err := getFileExtension(response.Header.Get("Content-Type"))
			if err != nil {
				return nil, err
			}
			img.Format = model.ImageFormat(format)

			var w, h string
			switch img.Format {
			case "jpg", "png", "jpeg", "gif":
				w, h, err = image.Dimensions(response.Body)
				if err != nil {
					log.Infof("Error parsing image resolution for %s: %v\n", imgURL.String(), err)
					continue
				}
			case "svg":
				w, h, err = image.SVGDimensions(response.Body)
				if err != nil {
					log.Infof("Error parsing svg image resolution for %s: %v\n", imgURL.String(), err)
					continue
				}
			case "webp":
				w, h, err = image.WebpDimensions(response.Body)
				if err != nil {
					log.Infof("Error parsing webp image resolution for %s: %v\n", imgURL.String(), err)
					continue
				}
			}

			img.Width = w
			img.Height = h
		case "alt":
			img.AltText = attr.Val
		case "title":
			img.Title = attr.Val
		}
	}

	return img, nil
}

func resolveURL(baseURL string, relativeURL *urllib.URL) *urllib.URL {
	base, err := urllib.Parse(baseURL)
	if err != nil {
		log.Infof("Error parsing base URL %s: %v\n", baseURL, err)
		return relativeURL
	}
	return base.ResolveReference(relativeURL)
}

func saveImage(ctx context.Context, imagesFolderName string, db db.DB, img *model.Image) error {
	resp, err := http.Get(img.SourceURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}

	// Create the file in the image directory
	filePath := filepath.Join("apps/imagecrawler/"+imagesFolderName, img.Filename+"."+string(img.Format))
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy the image content to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	_, err = db.SaveImage(ctx, img)
	if err != nil {
		return err
	}

	return nil
}

func getFileExtension(contentType string) (string, error) {
	switch {
	case strings.HasPrefix(contentType, "image/jpeg"):
		return "jpg", nil
	case strings.HasPrefix(contentType, "image/jpg"):
		return "jpg", nil
	case strings.HasPrefix(contentType, "image/png"):
		return "png", nil
	case strings.HasPrefix(contentType, "image/gif"):
		return "gif", nil
	case strings.HasPrefix(contentType, "image/svg"):
		return "svg", nil
	case strings.HasPrefix(contentType, "image/webp"):
		return "webp", nil
	default:
		return "", errors.New("invalid format")
	}
}
