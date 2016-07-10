package downloader

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"

	"golang.org/x/net/context"

	"github.com/ftloc/exception"
	"github.com/jpicht/logger"
)

type Downloader interface {
	Download(context.Context, *url.URL) Downloaded
}

type downloader struct {
}

func (d *downloader) Download(ctx context.Context, u *url.URL) Downloaded {
	log := logger.MustFromContext(ctx)
	switch u.Scheme {
	case "http", "https":
		log.WithData("url", u).Infof("HTTP GET")

		resp, err := http.Get(u.String())
		exception.ThrowOnError(err, err)
		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)
		exception.ThrowOnError(err, err)

		_, name := filepath.Split(u.Path)
		return NewDownloaded(name, data)
	default:
		exception.Throw(fmt.Errorf("Unsupported URL scheme: %s", u.Scheme))
	}
	return nil
}

func NewDownloader() Downloader {
	return &downloader{}
}
