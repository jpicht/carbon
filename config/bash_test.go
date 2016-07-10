package config

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/ftloc/exception"
)

func TestNewBashTasker(t *testing.T) {
	_ = NewBashTasker("echo ok")
}

func TestRunBashTasker(t *testing.T) {
	ta := NewBashTasker("echo ok")
	exception.Try(func() {
		ta.Run(context.Background())
	}).CatchAll(func(interface{}) {
		t.Fail()
	}).Finally(func() {})
}

func TestRunBashTasker2(t *testing.T) {
	ta := NewBashTasker("exit 1")
	flew := false
	exception.Try(func() {
		ta.Run(context.Background())
	}).CatchAll(func(interface{}) {
		flew = true
	}).Finally(func() {})

	if !flew {
		t.Fail()
	}
}
