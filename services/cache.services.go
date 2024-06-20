package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const dateFormat string = "20060102"
const timestampFormat string = "200601021504"
const cacheBaseDir string = "./tmp/cache/"

type Cache struct {
	Name  string
	Start time.Time
	End   time.Time
}

func (cache *Cache) Load() (data []byte, err error) {
	times := cache.getCachedTimes()
	if len(times) == 0 {
		return data, errors.New("error: there are no such caches")
	}
	data, err = os.ReadFile(cache.getPath(times[0]))
	if err != nil {
		return data, err
	}
	return data, nil
}

func (cache *Cache) Write(text []byte) (err error) {
	if _, err = os.Stat(cache.getDir()); os.IsNotExist(err) {
		os.MkdirAll(cache.getDir(), 0700)
	}
	err = os.WriteFile(cache.getPath(time.Now()), text, 0644)
	cache.cleanup()

	return err
}

func (cache *Cache) IsValid() bool {
	times := cache.getCachedTimes()
	// cache is invalid if there are no saved files
	if len(times) == 0 {
		return false
	}
	// building current time from string to ignore timezones
	now, _ := time.Parse(timestampFormat, time.Now().Format(timestampFormat))
	// cache is valid if it is less than one hour old
	return times[0].After(now.Add(-1 * time.Hour))
}

func (cache *Cache) getDir() (dir string) {
	return filepath.Join(cacheBaseDir, filepath.Base(cache.Name))
}

func (cache *Cache) getPathStrTime(timestamp string) (path string) {
	file := fmt.Sprintf(
		"%s-%s-%s-%s",
		cache.Name,
		cache.Start.Format(dateFormat),
		cache.End.Format(dateFormat),
		timestamp,
	)
	return filepath.Join(cache.getDir(), filepath.Base(file))
}

func (cache *Cache) getPath(timestamp time.Time) (path string) {
	return cache.getPathStrTime(timestamp.Format(timestampFormat))
}

func (cache *Cache) getCachedTimes() (timestamps []time.Time) {
	matches, _ := filepath.Glob(cache.getPathStrTime("*"))

	for _, file := range matches {
		timestamp, err := time.Parse(timestampFormat, strings.Split(filepath.Base(file), "-")[3])
		if err == nil {
			timestamps = append(timestamps, timestamp)
		}
	}

	// sorts timestamps with the most recent being first
	slices.SortFunc(timestamps, func(a, b time.Time) int {
		if a.Before(b) {
			return 1
		}
		if b.Before(a) {
			return -1
		}
		return 0
	})

	return timestamps
}

func (cache *Cache) cleanup() {
	times := cache.getCachedTimes()
	if len(times) < 2 {
		return
	}

	for _, oldTime := range times[1:] {
		os.Remove(cache.getPath(oldTime))
	}
}
