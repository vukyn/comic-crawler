package service

import (
	"fmt"
	"os"

	"github.com/go-pdf/fpdf"
)

func ImagesToPDF(folderPath string, pdfPath string, pdfName string) error {
	// Read all files in folder
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return err
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	for _, f := range files {
		info, err := f.Info()
		if err != nil {
			return err // in case of file removed or renamed
		}
		if info.IsDir() {
			continue
		}
		pdf.AddPage()

		imgPath := fmt.Sprintf("%s/%s", folderPath, f.Name())

		// Get image size
		// img, err := imgio.Open(imgPath)
		// if err != nil {
		// 	return err
		// }
		// w, h := img.Bounds().Dx(), img.Bounds().Dy()

		// Render image to PDF
		opts := fpdf.ImageOptions{ImageType: "JPG", ReadDpi: true}
		pdf.ImageOptions(imgPath, 0, 0, 0, 0, false, opts, 0, "")
	}

	output := fmt.Sprintf("%s/%s.pdf", pdfPath, pdfName)
	if err := pdf.OutputFileAndClose(output); err != nil {
		return err
	}

	return nil
}
