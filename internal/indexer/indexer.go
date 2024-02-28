package indexer

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/waynezhang/foto/internal/config"
	"github.com/waynezhang/foto/internal/images"
)

type Section struct {
	Title     string
	Text      template.HTML
	Slug      string
	Folder    string
	Ascending bool
	ImageSets []ImageSet
}

type ImageSize images.ImageSize

type ImageSet struct {
	FileName      string
	ThumbnailSize ImageSize
	OriginalSize  ImageSize
}

func Build(metadata []config.SectionMetadata, option config.ExtractOption) ([]Section, error) {
	sections := []Section{}
	slugs := map[string]bool{}

	for _, val := range metadata {
		slug := val.Slug
		if !validSlug(slug) {
			return nil, fmt.Errorf("Slug \"%s\" is invalid. Only letters([a-zA-Z]), numbers([09-]), underscore(_) and hyphen(-) can be used.", slug)
		}
		if slugs[slug] {
			return nil, fmt.Errorf("Slug \"%s\" already exists. Slug needs to be unique.", slug)
		}

		log.Debug().Msgf("Extacting section [%s][/%s] %s", val.Title, val.Slug, val.Folder)
		s := Section{
			Title:     val.Title,
			Text:      val.Text,
			Slug:      slug,
			Folder:    val.Folder,
			Ascending: val.Ascending,
			ImageSets: buildImageSets(val.Folder, val.Ascending, option),
		}
		slugs[slug] = true

		sections = append(sections, s)
	}

	return sections, nil
}

func buildImageSets(folder string, ascending bool, option config.ExtractOption) []ImageSet {
	sets := []ImageSet{}

	wg := &sync.WaitGroup{}
	mutext := &sync.Mutex{}

	_ = filepath.WalkDir(folder, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			log.Warn().Msgf("Failed to extract info from %s (%v)", path, err)
			return nil
		}
		if info.IsDir() || !images.IsPhotoSupported(path) {
			return nil
		}

		wg.Add(1)

		go func(src string) {
			defer wg.Done()

			s, err := buildImageSet(src, option)
			if s != nil {
				mutext.Lock()
				sets = append(sets, *s)
				mutext.Unlock()
			} else {
				log.Warn().Msgf("Failed to extract info from %s (%v)", src, err)
			}
		}(path)

		return nil
	})
	wg.Wait()

	sort.SliceStable(sets, func(i, j int) bool {
		if ascending {
			return sets[i].FileName < sets[j].FileName
		} else {
			return sets[i].FileName > sets[j].FileName
		}
	})

	return sets
}

func buildImageSet(path string, option config.ExtractOption) (*ImageSet, error) {
	imageSize, err := images.GetPhotoSize(path)
	if err != nil {
		return nil, err
	}

	thumbnailWidth := option.ThumbnailWidth
	thumbnailHeight := images.AspectedHeight(*imageSize, thumbnailWidth)

	originalWidth := option.OriginalWidth
	originalHeight := images.AspectedHeight(*imageSize, originalWidth)

	return &ImageSet{
		FileName: filepath.Base(path),
		ThumbnailSize: ImageSize{
			thumbnailWidth,
			thumbnailHeight,
		},
		OriginalSize: ImageSize{
			originalWidth,
			originalHeight,
		},
	}, nil
}

func validSlug(slug string) bool {
	matched, _ := regexp.MatchString("^[a-zA-Z0-9-_]+$", slug)
	return matched
}
