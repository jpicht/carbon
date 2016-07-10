package sourcer

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ftloc/exception"
	"github.com/ftloc/temp"
	"github.com/jpicht/carbon/config"
	"github.com/jpicht/carbon/downloader"
	"github.com/jpicht/logger"
	"golang.org/x/net/context"
)

type Sourcer interface {
	Source(config.Source) temp.Dir
}

type sourcer struct {
	context    context.Context
	downloader downloader.Downloader
}

func NewSourcer(ctx context.Context, dir string) Sourcer {
	return &sourcer{
		context:    ctx,
		downloader: downloader.NewCache(ctx, dir, downloader.NewDownloader()),
	}
}

func (s *sourcer) Source(src config.Source) temp.Dir {
	log := logger.MustFromContext(s.context)
	log.Infof("Sourcing %#v", src)

	u, err := url.Parse(src.Url)
	exception.ThrowOnError(err, err)
	switch src.Type {
	case "archive":
		d := s.downloader.Download(s.context, u)
		return s.extract(d)
	default:
		exception.Throw(fmt.Errorf("Unknown source type: %s", src.Type))
	}
	return nil
}

func (s *sourcer) extract(d downloader.Downloaded) temp.Dir {
	log := logger.MustFromContext(s.context).WithData("filename", d.Filename())
	extractors := map[string]func() temp.Dir{
		".tar":      func() temp.Dir { return s.untar("", d) },
		".tar.gz":   func() temp.Dir { return s.untar("-z", d) },
		".tar.bz2":  func() temp.Dir { return s.untar("-j", d) },
		".tar.xz":   func() temp.Dir { return s.untar("-J", d) },
		".tar.lzma": func() temp.Dir { return s.untar("--lzma", d) },
		".zip": func() temp.Dir {
			exception.Throw("Not implemented.")
			return nil
		},
	}
	for suf, fn := range extractors {
		if strings.HasSuffix(d.Filename(), suf) {
			log.Debugf("suf %s matches", suf)
			td := fn()
			files, err := ioutil.ReadDir(td.Name())
			exception.ThrowOnError(err, err)
			if len(files) == 1 {
				return &subDir{
					pDir:   td,
					subDir: filepath.Join(td.Name(), files[0].Name()),
				}
			}
			return td
		}
	}
	exception.Throw(fmt.Errorf("Unknown file type: %s", d.Filename()))
	return nil
}

type subDir struct {
	pDir   temp.Dir
	subDir string
}

func (d *subDir) Name() string {
	return d.subDir
}

func (d *subDir) Delete() {
	d.pDir.Delete()
}

func (s *sourcer) untar(c string, d downloader.Downloaded) temp.Dir {
	log := logger.MustFromContext(s.context).WithData("filename", d.Filename())
	args := []string{}
	if len(c) > 0 {
		args = append(args, c)
	}
	tarfile := temp.NewFile().WithSuffix(d.Filename()).Create()
	tarfile.Write(d.Data())
	defer tarfile.Delete()

	dest := temp.NewDir().Create()
	args = append(args, "-x", "-f", tarfile.Name(), "-C", dest.Name())
	cmd := exec.Command("tar", args...)
	log.Infof("%#v", cmd)

	pipe, err := cmd.StderrPipe()
	exception.ThrowOnError(err, err)
	sc := bufio.NewScanner(pipe)
	if err := cmd.Run(); err != nil {
		for sc.Scan() {
			log.Error(strings.TrimSpace(sc.Text()))
		}
		exception.Throw("Extracting tar file failed.")
	}

	return dest
}
