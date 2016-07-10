package main

import (
	"flag"
	"os/user"

	"golang.org/x/net/context"

	"github.com/jpicht/carbon/config"
	"github.com/jpicht/carbon/sourcer"
	"github.com/jpicht/logger"
)

func main() {
	log := logger.NewStderrLogger()
	file := flag.String("file", "", "Config file")
	flag.Parse()

	log.Infof("Config file '%s'", *file)

	ctx := log.Context(context.Background())
	recipe := config.LoadFile(ctx, *file)

	usr, err := user.Current()
	if err != nil {
		log.Errorf("Cannot get current user info: %s", err.Error())
	}

	s := sourcer.NewSourcer(ctx, usr.HomeDir+"/.carbon/cache")
	destdir := s.Source(recipe.Source)
	defer destdir.Delete()

	log.Infof("Temporary directory: %s", destdir.Name())
	for _, task := range recipe.Tasks {
		task.Run(log.Context(context.Background()), &config.BuildConfig{
			TempDirectory:   destdir.Name(),
			TargetDirectory: usr.HomeDir + "/.carbon/prefix/" + recipe.Name,
		})
	}
}
