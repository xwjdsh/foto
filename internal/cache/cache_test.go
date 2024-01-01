package cache

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/waynezhang/foto/internal/constants"
	"github.com/waynezhang/foto/internal/files"
	"github.com/waynezhang/foto/internal/test"
)

func TestCache(t *testing.T) {
	dirName, err := os.MkdirTemp("", "foto-cache")
	assert.Nil(t, err)
	defer os.RemoveAll(dirName)

	instance = New(dirName)

	assert.Equal(t, dirName, instance.directoryName)

	img := instance.CachedImage(test.Testfile, 640)
	assert.Nil(t, img)

	instance.AddImage(test.Testfile, 640, test.ThumbnailFile)
	img = instance.CachedImage(test.Testfile, 640)
	expectedPath := fmt.Sprintf("%s/%s-640", dirName, test.ExpectedChecksum)
	assert.Equal(t, expectedPath, *img)

	resizedChecksum, _ := files.Checksum(expectedPath)
	assert.Equal(t, test.ExpectedThubmnailChecksum, *resizedChecksum)

	instance.Clear()
	assert.NoFileExists(t, dirName)
}

func TestShared(t *testing.T) {
	instance := Shared()
	assert.Equal(t, constants.CacheDirectoryName, instance.directoryName)
}

func TestImagePath(t *testing.T) {
	instance := New("some-path")
	path := instance.imagePath("some-checksum", 200)
	assert.Equal(t, "some-path/some-checksum-200", path)
}
