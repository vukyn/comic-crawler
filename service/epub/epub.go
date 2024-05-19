package epub

import (
	"fmt"
	"os"
	"strings"

	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"
	gub "github.com/go-shiori/go-epub"
	"github.com/vukyn/kuery/file"
	"github.com/vukyn/kuery/log"
	"github.com/vukyn/kuery/query/v2"
)

type EpubOption struct {
	Title  string
	Author string
	Cover  string
	RTL    bool
}

func ImagesToEPUB(folderPath, filePath, fileName string, opt EpubOption) error {
	// Read all files in folder
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return err
	}

	// Set default
	title := opt.Title
	if title == "" {
		title = fileName
	}

	// init EPUB
	e, err := gub.NewEpub(title)
	if err != nil {
		return err
	}
	e.SetAuthor(opt.Author)

	// Set RTL
	if opt.RTL {
		e.SetPpd("rtl")
	}

	// Load template page
	comicPage, err := os.ReadFile("service/epub/template/comic_page.html")
	if err != nil {
		return err
	}

	// Add css to EPUB
	internalCSS, err := e.AddCSS("service/epub/template/stylesheet.css", "stylesheet.css")
	if err != nil {
		return err
	}

	// Update cover css
	coverCSS, err := e.AddCSS("service/epub/template/cover.css", "")
	if err != nil {
		return err
	}

	// Add image cover to EPUB
	coverImg, err := e.AddImage(opt.Cover, "cover.jpg")
	if err != nil {
		return err
	}

	// Add cover to EPUB
	if err = e.SetCover(coverImg, coverCSS); err != nil {
		return err
	}

	for i := range files {
		info, ok := query.FindFunc(files, func(f os.DirEntry) bool {
			return f.Name() == fmt.Sprintf("%d.jpg", i+1)
		})
		if !ok {
			log.Warnf("Missing image %d.jpg", i+1)
			continue
		}
		if info.IsDir() {
			continue
		}

		imgPath := fmt.Sprintf("%s/%s", folderPath, info.Name())

		// Check if image need to be rotated
		isRotated := false
		img, err := imgio.Open(imgPath)
		if err != nil {
			return err
		}
		w, h := img.Bounds().Dx(), img.Bounds().Dy()
		if w > h {
			isRotated = true
			img = transform.Rotate(img, -90, &transform.RotationOptions{ResizeBounds: true})
			imgPath = fmt.Sprintf("%s/%s", folderPath, strings.Split(info.Name(), ".")[0]+"_rotated.jpg")
			if err := imgio.Save(imgPath, img, imgio.PNGEncoder()); err != nil {
				return err
			}
		}

		// Add image to EPUB
		imgSrc, err := e.AddImage(imgPath, info.Name())
		if err != nil {
			return err
		}

		// Add section to EPUB
		htmlPage := string(comicPage)
		htmlPage = strings.ReplaceAll(htmlPage, "[[img]]", imgSrc)
		if isRotated {
			htmlPage = strings.ReplaceAll(htmlPage, "[[rotate]]", "-rotate")
		} else {
			htmlPage = strings.ReplaceAll(htmlPage, "[[rotate]]", "")
		}
		if _, err := e.AddSection(htmlPage, fmt.Sprintf("%v - Part %v", opt.Title, i+1), fmt.Sprintf("part%d", i+1), internalCSS); err != nil {
			return err
		}
	}

	if err := file.CreateFilePath(fmt.Sprintf("%s/epub/", filePath)); err != nil {
		return err
	}
	if err := e.Write(fmt.Sprintf("%s/epub/%s.epub", filePath, fileName)); err != nil {
		return err
	}

	// Clean up rotated image
	files, err = os.ReadDir(folderPath)
	if err != nil {
		return err
	}
	for _, f := range files {
		info, err := f.Info()
		if err != nil {
			return err // in case of file removed or renamed
		}
		if info.IsDir() {
			continue
		}
		if strings.Contains(info.Name(), "_rotated") {
			if err := os.Remove(fmt.Sprintf("%s/%s", folderPath, info.Name())); err != nil {
				log.Errorf("Error removing rotated image: %v", err)
			}
			continue
		}
	}

	return nil
}
