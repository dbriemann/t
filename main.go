package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/parnurzeal/gorequest"
	"github.com/shibukawa/configdir"
)

const (
	configFile = "db.json"
)

var (
	configDir configdir.ConfigDir
	config    *configdir.Config
	db        = DB{}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func help() {
	fmt.Println("")
	fmt.Println("USAGE:")
	fmt.Println(" Set a timer (overwrites existing)")
	fmt.Println(" ---------------------------------")
	fmt.Println("  t name duration [target]")
	fmt.Println("")
	fmt.Println(" Start a named timer")
	fmt.Println(" -------------------")
	fmt.Println("  t name")
	fmt.Println("")
	fmt.Println(" Rename a timer")
	fmt.Println(" --------------")
	fmt.Println("  t name = newname")
	fmt.Println("")
	fmt.Println(" Delete a named timer")
	fmt.Println(" -------------------")
	fmt.Println("  t name del")
	fmt.Println("")
	fmt.Println(" Start a custom timer")
	fmt.Println(" --------------------")
	fmt.Println("  t duration")
	fmt.Println("")
	fmt.Println(" Parameters")
	fmt.Println(" ----------")
	fmt.Println("    name / newname: no spaces")
	fmt.Println("    duration: countdown time, format '1h23m1s'")
	fmt.Println("    target: file or url to be opened when timer triggers, absolute path")
	fmt.Println("    id: number in table")
	fmt.Println("")
}

func list() {
	fmt.Println("")
	fmt.Println("Saved timers:")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Name", "Countdown", "Target", "Used"})

	for i, a := range db.Timers {
		row := make([]string, 5)
		row[0] = strconv.Itoa(i + 1)

		row[1] = a.Name
		row[2] = a.Countdown.String()
		tlen := len(a.Target)
		if tlen > 30 {
			row[3] = a.Target[:14] + ".." + a.Target[tlen-14:]
		} else {
			row[3] = a.Target
		}
		row[4] = strconv.Itoa(int(a.Used))

		table.Append(row)
	}

	table.Render() // Send output
	fmt.Println("")
}

func fetchDailyPhoto() (url string) {
	url = "https://www.google.com/search?q=something+went+wrong&oq=something+went+wrong"
	type entry struct {
		// Omit everything but the URL
		URL string `json:"url"`
	}
	type response struct {
		Entries []entry `json:"images"`
	}
	reply := response{}
	_, _, errs := gorequest.New().Get("https://www.bing.com/HPImageArchive.aspx?format=js&idx=0&n=8").EndStruct(&reply)
	if errs == nil {
		r := rand.Intn(len(reply.Entries))
		url = "https://bing.com" + reply.Entries[r].URL
	}
	return
}

func platformOpen(target string) error {
	cmd := exec.Command(platformOpenCmd, target)
	return cmd.Run()
}

func validateTarget(t string) bool {
	// 1. Test if t is a valid URL.
	_, err := url.ParseRequestURI(t)
	if err == nil {
		return true
	}

	// 2. Test if t is an existing file system path.
	_, err = os.Stat(t)
	return os.IsExist(err)
}

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

	args := os.Args[1:]

	if len(args) == 0 {
		help()
		list()
		os.Exit(0)
	} else if len(args) > 3 {
		fmt.Println("too many arguments")
		os.Exit(1)
	} else {
		first := args[0]
		// 1. Is first input a duration?
		duration, err := time.ParseDuration(first)
		if err == nil {
			// Run an unnamed timer.
			t := Timer{
				Countdown: duration,
				Name:      "custom",
				Target:    "",
			}
			t.run()
			os.Exit(0)
		} else {
			// We assume that first input is a name for now.
			if len(args) == 1 {
				// Search if timer with given name exists and if so run it.
				ran := false
				for i := 0; i < len(db.Timers); i++ {
					if db.Timers[i].Name == first {
						db.Timers[i].run()
						db.Timers[i].Used++
						db.save()
						os.Exit(0)
					}
				}
				if !ran {
					fmt.Printf("timer with name %s does not exist\n", args[0])
					os.Exit(1)
				}
			} else {
				second := args[1]
				duration, err = time.ParseDuration(second)
				if err != nil {
					// If not a duration it only can be "del" or "=" command.
					if second == "del" {
						fmt.Printf("deleting timer '%s'\n", first)
						db.delete(first)
						list()
						os.Exit(0)
					} else if second == "=" {
						if len(args) < 3 {
							fmt.Println("missing newname parameter")
							os.Exit(1)
						}
						third := args[2]
						// Rename action.
						if ok := db.renameTimer(first, third); !ok {
							fmt.Println("no timer with that name")
							os.Exit(1)
						} else {
							db.save()
							list()
							os.Exit(0)
						}
					} else {
						fmt.Printf("unknown command: %s\n", second)
						os.Exit(1)
					}
				} else {
					// We have a valid duration here.

					// If there is no optional target just use "".
					target := ""
					if len(args) == 3 {
						target = args[2]
						valid := validateTarget(target)
						if !valid {
							fmt.Println("target is not a valid URI and not a valid file path")
							fmt.Println("using default target..")
							target = ""
						}
					}

					// Add to DB
					db.setTimer(first, target, duration)
					db.save()
					list()
					os.Exit(0)
				}
			}
		}
	}
}
