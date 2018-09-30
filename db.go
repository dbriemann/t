package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/shibukawa/configdir"
)

type Timer struct {
	Countdown time.Duration
	Name      string
	Target    string
	Used      uint32
}

func (t *Timer) run() {
	fmt.Printf("running timer '%s' %s %s\n", t.Name, t.Countdown.String(), t.Target)
	tim := time.NewTimer(t.Countdown)
	<-tim.C
	target := t.Target
	if target == "" {
		target = fetchDailyPhoto()
	}
	err := platformOpen(target)
	if err != nil {
		fmt.Printf("failure executing open command: %s\n", err.Error())
		os.Exit(1)
	}
}

type DB struct {
	Timers []Timer
}

func (db *DB) save() error {
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

func (db *DB) renameTimer(name, newname string) bool {
	for i := 0; i < len(db.Timers); i++ {
		if db.Timers[i].Name == name {
			db.Timers[i].Name = newname
			return true
		}
	}
	return false
}

func (db *DB) delete(name string) {
	el := -1
	for i := 0; i < len(db.Timers); i++ {
		if db.Timers[i].Name == name {
			el = i
			break
		}
	}

	if el >= 0 {
		db.Timers = append(db.Timers[:el], db.Timers[el+1:]...)
		db.save()
	}
}

func (db *DB) setTimer(name, target string, dur time.Duration) {
	t := &Timer{}
	found := false

	// Check if timer name is already present.
	for i := 0; i < len(db.Timers); i++ {
		if db.Timers[i].Name == name {
			t = &db.Timers[i]
			found = true
			break
		}
	}

	t.Name = name
	t.Target = target
	t.Countdown = dur

	if !found {
		db.Timers = append(db.Timers, *t)
	}
}
