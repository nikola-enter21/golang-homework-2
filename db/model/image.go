package model

type Image struct {
	Name      string      `db:"name"`
	Filename  string      `db:"filename"`
	AltText   string      `db:"alt_text"`
	Title     string      `db:"title"`
	Width     string      `db:"width"`
	Height    string      `db:"height"`
	Format    ImageFormat `db:"format"`
	SourceURL string      `db:"source_url"`
}

// ImageFormat represents the format of an image
type ImageFormat string

const (
	FormatJPG  ImageFormat = "jpg"
	FormatJPEG ImageFormat = "jpeg"
	FormatPNG  ImageFormat = "png"
	FormatGIF  ImageFormat = "gif"
	FormatSVG  ImageFormat = "svg"
	FormatWEBP ImageFormat = "webp"
)

type ListImagesFilters struct {
	Format ImageFormat `json:"format"`
}

func ImageFormatValues() []ImageFormat {
	return []ImageFormat{FormatPNG, FormatGIF, FormatSVG, FormatJPG, FormatJPEG, FormatWEBP}
}

func ImagesCountPerFormat(images []*Image) map[ImageFormat]uint32 {
	m := make(map[ImageFormat]uint32)

	for _, img := range images {
		m[img.Format]++
	}

	return m
}
