package WaterMark

import (
	"image/color"
	"os"
	"testing"
)

func TestWaterMark(t *testing.T) {
	demoTest := WaterMark{
		Source:    "./test.png",
		Text:      "tulian",
		TextColor: color.RGBA{0, 0, 255, 255},
		Scale:     1.0,
		Format:    "png",
	}

	outputFile, _ := os.Create("./mark_test.png")
	defer outputFile.Close()

	_, err := demoTest.CreateWaterMark(outputFile)

	if err != nil {
		t.Error(err.Error())
	}
}
