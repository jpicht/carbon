package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/ftloc/exception"
	"golang.org/x/net/context"
)

func Load(ctx context.Context, rawJson []byte) *Recipe {
	r := &Recipe{context: ctx}
	err := json.Unmarshal(rawJson, r)
	exception.ThrowOnError(err, err)
	return r
}

func LoadFile(ctx context.Context, fileName string) *Recipe {
	f, err := os.Open(fileName)
	exception.ThrowOnError(err, err)
	d, err := ioutil.ReadAll(f)
	exception.ThrowOnError(err, err)
	return Load(ctx, d)
}
