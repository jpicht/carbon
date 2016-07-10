package downloader

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"unicode"

	"github.com/ftloc/exception"
	"github.com/jpicht/logger"
	"golang.org/x/net/context"
)

type cacher struct {
	loader   Downloader
	cacheDir string
}

func NewCache(ctx context.Context, dir string, l Downloader) Downloader {
	log := logger.MustFromContext(ctx)
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			log.Infof(dir + " does not exist")
			err := os.MkdirAll(dir, 0755)
			exception.ThrowOnError(err, err)
		} else {
			exception.ThrowOnError(err, err)
		}
	}
	return &cacher{
		loader:   l,
		cacheDir: dir,
	}
}

type metadata struct {
	Filename string `json:"filename"`
}

func (c *cacher) Download(ctx context.Context, u *url.URL) Downloaded {
	cf := c.cacheFile(u)
	log := logger.MustFromContext(ctx).WithData("cache-file", cf)
	if c.inCache(u) {
		var result Downloaded = nil
		exception.Try(func() {
			f, err := os.Open(cf)
			exception.ThrowOnError(err, err)
			defer f.Close()
			data, err := ioutil.ReadAll(f)
			exception.ThrowOnError(err, err)

			f2, err := os.Open(c.metaFile(u))
			exception.ThrowOnError(err, err)
			defer f2.Close()
			meta, err := ioutil.ReadAll(f2)
			exception.ThrowOnError(err, err)

			dec_meta := metadata{}
			err = json.Unmarshal(meta, &dec_meta)
			exception.ThrowOnError(err, err)

			log.Infof("Served from cache.")
			result = NewDownloaded(dec_meta.Filename, data)
		}).CatchAll(func(i interface{}) {
			log.Warningf("Cache invalid: %#v", i)
		}).Finally(func() {})
		if result != nil {
			return result
		}
	}
	data := c.loader.Download(ctx, u)
	err := ioutil.WriteFile(cf, data.Data(), 0644)
	meta, err := json.Marshal(&metadata{data.Filename()})
	exception.ThrowOnError(err, err)
	err = ioutil.WriteFile(c.metaFile(u), meta, 0644)
	exception.ThrowOnError(err, err)
	return data
}

func (c *cacher) inCache(u *url.URL) bool {
	f := c.cacheFile(u)
	if _, err := os.Stat(f); err != nil {
		if os.IsNotExist(err) {
			return false
		}
		exception.ThrowOnError(err, err)
	}
	return true
}

func (c *cacher) cacheFile(u *url.URL) string {
	uString := u.String()
	oString := "file_"
	mangle(uString, &oString)
	return c.cacheDir + "/" + oString
}

func (c *cacher) metaFile(u *url.URL) string {
	uString := u.String()
	oString := "meta_"
	mangle(uString, &oString)
	return c.cacheDir + "/" + oString + ".json"
}

func mangle(s string, o *string) {
	for i := 0; i < len(s); i++ {
		r := rune(s[i])
		if s[i] != '.' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			*o += "_"
		} else {
			*o += string(s[i])
		}
	}
}
