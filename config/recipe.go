package config

import (
	"encoding/json"
	"fmt"

	"github.com/jpicht/logger"

	"golang.org/x/net/context"
)

type Recipe struct {
	Name    string
	Source  Source
	Tasks   []Tasker
	context context.Context
}

type Source struct {
	Type string  `json:"type"`
	Url  string  `json:"url"`
	Hash *string `json:"hash"`
}

func (r *Recipe) UnmarshalJSON(data []byte) error {
	log := logger.MustFromContext(r.context)

	// first unmarshal into a helper structure
	temp := &struct {
		Name   string            `json:"name"`
		Source Source            `json:"source"`
		Tasks  []json.RawMessage `json:"tasks"`
	}{}

	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}

	// copy data, while decoding tasks
	r.Name = temp.Name
	r.Source = temp.Source

	log.Infof("Loading recipe for %s", r.Name)

	t := &struct {
		Type string `json:"type"`
	}{}
	for _, tt := range temp.Tasks {
		var tmpTask Tasker

		err = json.Unmarshal(tt, t)
		if nil != err {
			return err
		}

		switch t.Type {
		case "bash":
			tmpTask = &BashTasker{}
		case "make":
			tmpTask = &MakeTasker{}
		default:
			return fmt.Errorf("Unknown task type: %s", t.Type)
		}

		err = json.Unmarshal(tt, tmpTask)
		if nil != err {
			return err
		}

		log.Infof("Task: %#v", tmpTask)
		r.Tasks = append(r.Tasks, tmpTask)
	}

	return nil
}

type taskTypeIdentifier struct {
	Type string `json:"type"`
}
