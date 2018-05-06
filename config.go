package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/shibukawa/configdir"
)

type Alarm struct {
	Countdown time.Duration
	Name      string
	RepeatX   int
}

type DB struct {
	Alarms []Alarm
}

func (db *DB) Save(path string) error {
	data, err := json.Marshal(db)
	if err != nil {
		fmt.Println("Error serializing db:", err.Error())
		return err
	}

	if config == nil {
		all := configDir.QueryFolders(configdir.Global)
		config = all[0]
	}

	err = config.WriteFile(configFile, data)
	if err != nil {
		fmt.Println("Error writing db:", err.Error())
		return err
	}

	return nil
}
