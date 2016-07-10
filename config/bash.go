package config

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/ftloc/exception"
	"github.com/jpicht/logger"
	"golang.org/x/net/context"
)

type BashTasker struct {
	Script            string
	ExpectedExitCodes []int
}

func (b *BashTasker) Run(ctx context.Context, c *BuildConfig) {
	script := strings.Replace(b.Script, "{{TargetDirectory}}", c.TargetDirectory, -1)

	cur, err := os.Getwd()
	exception.ThrowOnError(err, err)
	defer func() {
		err := os.Chdir(cur)
		exception.ThrowOnError(err, err)
	}()
	err = os.Chdir(c.TempDirectory)
	exception.ThrowOnError(err, err)

	log := logger.MustFromContext(ctx)
	log.WithData("script", script).Infof("running")

	cmd := exec.Command("bash", "-c", script)
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
	if eErr, ok := err.(*exec.ExitError); ok {
		if status, ok := eErr.Sys().(syscall.WaitStatus); ok {
			for _, allowed := range b.ExpectedExitCodes {
				if allowed == status.ExitStatus() {
					return
				}
			}
		}
	}
	exception.ThrowOnError(err, err)
}

func (b *BashTasker) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type              string `json:"type"`
		Script            string `json:"script"`
		ExpectedExitCodes []int  `json:"expectedExitCodes"`
	}{
		Type:   "bash",
		Script: b.Script,
	})
}

func (b *BashTasker) UnmarshalJSON(data []byte) error {
	tmp := &struct {
		Script            string `json:"script"`
		ExpectedExitCodes []int  `json:"expectedExitCodes"`
	}{}

	err := json.Unmarshal(data, tmp)
	if nil != err {
		return err
	}

	b.Script = tmp.Script
	b.ExpectedExitCodes = tmp.ExpectedExitCodes
	return nil
}

func NewBashTasker(script string) Tasker {
	return &BashTasker{
		Script: script,
	}
}
