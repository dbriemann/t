package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/shibukawa/configdir"
)

const (
	configFile = "config.json"
)

var (
	configDir configdir.ConfigDir
	config    *configdir.Config
	db        = DB{}
)

func main() {
	configDir = configdir.New("", "t")
	config = configDir.QueryFolderContainsFile(configFile)

	if config != nil {
		data, err := config.ReadFile(configFile)
		if err != nil {
			fmt.Println("Could not read config file:", err.Error())
			os.Exit(1)
		}

		err = json.Unmarshal(data, &db)
		if err != nil {
			fmt.Println("Could not de-serialize config file:", err.Error())
			os.Exit(1)
		}
	}

}
