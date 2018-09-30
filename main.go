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

// TODO - restructure command switch case code to be prettier.
// TODO - use daily photo if no target is provided.

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
	fmt.Println(" Rename a timer")
	fmt.Println(" --------------")
	fmt.Println("  t name = newname")
	fmt.Println("")
	fmt.Println(" Start a named timer")
	fmt.Println(" -------------------")
	fmt.Println("  t name \n  t /id")
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
	} else {
		// Test input for shortcut.
		if args[0][0] == '/' {
			// Start a timer via shortcut if it exists.
			num, err := strconv.Atoi(args[0][1:])
			if err != nil {
				fmt.Printf("shortcut malformed: %s\n", err.Error())
				os.Exit(1)
			}
			num--
			if num >= len(db.Timers) || num < 0 {
				fmt.Printf("shortcut %d does not exist\n", num+1)
			} else {
				db.Timers[num].run()
				db.Timers[num].Used++
				db.save()
				os.Exit(0)
			}
		} else {
			// Test if there are 3 or more arguments. If so a timer is set / updated.
			if len(args) >= 3 {
				name := args[0]
				if args[1] == "=" {
					// Rename action.
					if ok := db.renameTimer(name, args[2]); !ok {
						fmt.Println("no timer with that name")
					} else {
						db.save()
					}
					list()
					os.Exit(1)
				}

				duration, err := time.ParseDuration(args[1])
				if err != nil {
					fmt.Printf("parameter duration is malformed: %s\n", err.Error())
					os.Exit(1)
				}
				target := args[2]
				valid := validateTarget(target)
				if !valid {
					fmt.Println("target is not a valid URI and not a valid file path")
					os.Exit(1)
				}

				// Add to DB
				db.setTimer(name, target, duration)
				db.save()
				list()
			} else {
				if len(args) == 2 && args[1] == "del" {
					fmt.Printf("deleting timer '%s'\n", args[0])
					db.delete(args[0])
					list()
					os.Exit(0)
				} else if len(args) > 1 && len(args) < 3 {
					fmt.Println("unkown input")
					os.Exit(1)
				}

				// Test if argument is a duration.
				dur, err := time.ParseDuration(args[0])
				if err == nil {
					t := Timer{
						Countdown: dur,
						Name:      "custom",
						Target:    fetchDailyPhoto(),
					}
					t.run()
					os.Exit(0)
				}

				// Search if timer with given name exists and if so run it.
				ran := false
				for i := 0; i < len(db.Timers); i++ {
					if db.Timers[i].Name == args[0] {
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
			}

		}
	}
}
