package main

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"strings"
	"syscall"
	"time"

	"github.com/chzyer/readline"
	"golang.org/x/crypto/ssh/terminal"
)

type command interface {
	// run a command with the given state, within the given group, returning the group the user should manipulate after this command is run.
	run(state *State, group string, args string, reader *bufio.Reader) string
}

type lsCommand struct{}
type entryCommand struct{}
type groupCommand struct{}

var commands = map[string]command{
	"ls":    lsCommand{},
	"group": groupCommand{},
	"entry": entryCommand{},
}

func createCompleter() *readline.PrefixCompleter {
	var cmdItems = make([]readline.PrefixCompleterInterface, len(commands)+1)
	i := 0
	for k := range commands {
		cmdItems[i] = readline.PcItem(k)
		i++
	}
	cmdItems[i] = readline.PcItem("exit")
	return readline.NewPrefixCompleter(cmdItems...)
}

func usage(w io.Writer) {
	io.WriteString(w, "Commands:\n")
	io.WriteString(w, createCompleter().Tree("    "))
}

func (cmd lsCommand) run(state *State, group string, args string, reader *bufio.Reader) string {
	if len(*state) == 0 {
		println("default:\n  <empty>")
	} else {
		for group, entries := range *state {
			fmt.Printf("Group: %s (%d entries)\n", group, len(entries))
			if len(entries) == 0 {
				println("  <empty>")
			} else {
				for _, e := range entries {
					println("  " + e.String())
				}
			}
		}
	}
	return group
}

func (cmd entryCommand) run(state *State, group string, args string, reader *bufio.Reader) string {
	newEntry := args
	if len(newEntry) > 0 {
		entries, _ := (*state)[group]
		entry, found := findEntryIn(entries, newEntry)
		if found {
			println(entry.String())
		} else {
			newEntryWanted := yesNoQuestion("Entry does not exist, do you want to create it? [y/n]: ", reader)
			if newEntryWanted {
				entry := createNewEntry(newEntry, reader)
				(*state)[group] = append(entries, entry)
			}
		}
	} else {
		println("Error: please provide a name for the entry")
	}
	return group
}

func (cmd groupCommand) run(state *State, group string, args string, reader *bufio.Reader) string {
	groupName := args
	if len(groupName) > 0 {
		_, groupExists := (*state)[groupName]
		if groupExists {
			return groupName
		}

		newGroupWanted := yesNoQuestion("Group does not exist, do you want to create it? [y/n]: ", reader)
		if newGroupWanted {
			for {
				ok := createNewGroup(state, groupName)
				if ok {
					return groupName
				}
				groupName = read(reader, "Please enter another name for the group: ")
			}
		}
	} else {
		println("Error: please provide a name for the group")
	}

	return group
}

func yesNoQuestion(question string, reader *bufio.Reader) bool {
	for {
		yn := strings.ToLower(read(reader, question))
		if len(yn) == 0 || yn == "y" {
			return true
		} else if yn == "n" {
			return false
		} else {
			println("Please answer y or n (no answer means y)")
		}
	}

}

func createNewGroup(state *State, name string) bool {
	if len(name) > 0 {
		_, ok := (*state)[name]
		if !ok {
			(*state)[name] = []LoginInfo{}
			return true
		}
		println("Error: group already exists")
	} else {
		println("Error: please provide a name for the group")
	}
	return false
}

func read(reader *bufio.Reader, prompt string) string {
	print(prompt)
	a, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(a)
}

func createNewEntry(name string, reader *bufio.Reader) (result LoginInfo) {
	username := read(reader, "Enter username: ")

	var URL string
	goodURL := false
	for !goodURL {
		URL = read(reader, "Enter URL: ")
		if len(URL) > 0 {
			_, err := url.Parse(URL)
			if err != nil {
				println("Invalid URL, please try again.")
			} else {
				goodURL = true
			}
		} else {
			goodURL = true // empty URL is ok
		}
	}

	description := read(reader, "Enter description: ")

	var password string
	answerAccepted := false
	for !answerAccepted {
		answer := strings.ToLower(read(reader, "Generate password? [y/n]: "))
		if len(answer) == 0 || answer == "y" {
			password = generatePassword()
			fmt.Printf("Generated password for %s!\n", name)
			fmt.Printf("To copy it to the clipboard, type 'cp %s'\n", name)
			answerAccepted = true
		} else if answer == "n" {
			for !answerAccepted {
				print("Please enter a password (at least 4 characters): ")
				password, err := terminal.ReadPassword(int(syscall.Stdin))
				println("")
				if err != nil {
					panic(err)
				}
				if len(password) < 4 {
					println("Password too short, please try again!")
				} else {
					answerAccepted = true
				}
			}
		} else {
			println("Please answer y or n (no answer means y)")
		}
	}

	result.Name = name
	result.Username = username
	result.URL = URL
	result.Password = password
	result.Description = description
	result.UpdatedAt = time.Now()

	return
}

func findEntryIn(entries []LoginInfo, name string) (LoginInfo, bool) {
	for _, e := range entries {
		if name == e.Name {
			return e, true
		}
	}
	return LoginInfo{}, false
}

func generatePassword() string {
	// TODO
	return ""
}
