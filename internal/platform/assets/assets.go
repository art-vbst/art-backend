package assets

import (
	"image"
	"log"
	"os"
)

type Assets struct {
	watermarkFile *os.File
	Watermark     image.Image
}

func Load() *Assets {
	var assets Assets

	file, err := os.Open("assets/watermark.png")
	if err != nil {
		log.Fatalf("unable to open watermark file: %v", err)
	}
	assets.watermarkFile = file

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatalf("unable to decode watermark image: %v", err)
	}
	assets.Watermark = img

	return &assets
}

func (a *Assets) Close() {
	if a.watermarkFile != nil {
		a.watermarkFile.Close()
	}
}
