package config

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	"github.com/ftloc/exception"
	"github.com/jpicht/logger"
	"golang.org/x/net/context"
)

type MakeTasker struct {
	Targets []string `json:"targets"`
}

func (b *MakeTasker) Run(ctx context.Context, c *BuildConfig) {
	log_ := logger.MustFromContext(ctx)
	for _, target := range b.Targets {
		cmd := exec.Command("make", target)
		log := log_.WithData("make", target)

		cur, err := os.Getwd()
		exception.ThrowOnError(err, err)
		defer func() {
			err := os.Chdir(cur)
			exception.ThrowOnError(err, err)
		}()
		err = os.Chdir(c.TempDirectory)
		exception.ThrowOnError(err, err)

		eout, err := cmd.StderrPipe()
		exception.ThrowOnError(err, err)
		sout, err := cmd.StdoutPipe()
		exception.ThrowOnError(err, err)

		se := bufio.NewScanner(eout)
		so := bufio.NewScanner(sout)
		go func() {
			for so.Scan() {
				log.Debug(strings.TrimSpace(so.Text()))
			}
		}()
		go func() {
			for se.Scan() {
				log.Warning(strings.TrimSpace(se.Text()))
			}
		}()
		err = cmd.Run()

		exception.ThrowOnError(err, err)
	}
}

func (b *MakeTasker) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type    string   `json:"type"`
		Targets []string `json:"targets"`
	}{
		Type:    "make",
		Targets: b.Targets,
	})
}

func (b *MakeTasker) UnmarshalJSON(data []byte) error {
	tmp := &struct {
		Targets []string `json:"targets"`
	}{}

	err := json.Unmarshal(data, tmp)
	if nil != err {
		return err
	}

	b.Targets = tmp.Targets
	return nil
}

func NewMakeTasker(targets []string) Tasker {
	return &MakeTasker{
		Targets: targets,
	}
}
