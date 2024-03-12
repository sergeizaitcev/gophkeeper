package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

// Commander описывает команду выполнения.
type Commander interface {
	run(parent string, args []string) error
	equal(name string) bool
	shortPrint()
}

// Command определяет основную команду выполнения.
type Command struct {
	Name        string      // Наименование команды.
	Description string      // Описание команды.
	Version     string      // Версия команды.
	Subcommands []Commander // Список подкоманд для выполнения.
}

// CommandGroup определяет командную группу.
type CommandGroup struct {
	Name        string      // Наименование командной группы.
	Description string      // Описание командной группы.
	Subcommands []Commander // Список подкоманд для выполнения.
}

// Subcommand определяет подкоманду выполнения.
type Subcommand struct {
	Name        string               // Наименование подкоманды.
	Description string               // Описание подкоманды.
	Flags       func(*flag.FlagSet)  // Функция регистрации флагов.
	Execute     func([]string) error // Функция выполнения подкоманды.
}

// Execute запускает выполнение команды.
func (cmd *Command) Execute() error {
	if len(os.Args) < 2 {
		cmd.usage()
		return nil
	}

	name := os.Args[1]
	if name == "version" {
		fmt.Printf("Version: %s\n", cmd.Version)
		return nil
	}

	sub, ok := lookup(name, cmd.Subcommands)
	if !ok {
		cmd.usage()
		return nil
	}

	return sub.run(cmd.Name, os.Args[2:])
}

func (group *CommandGroup) run(parent string, args []string) error {
	if len(args) == 0 {
		group.usage(parent)
		return nil
	}

	name := args[0]

	sub, ok := lookup(name, group.Subcommands)
	if !ok {
		group.usage(parent)
		return nil
	}

	return sub.run(parent+" "+group.Name, args[1:])
}

func (sub *Subcommand) run(parent string, args []string) error {
	fs := flag.NewFlagSet(sub.Name, flag.ContinueOnError)
	fs.Usage = func() {}

	if sub.Flags != nil {
		sub.Flags(fs)
	}
	if err := fs.Parse(args); err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			fmt.Println()
		}
		sub.usage(parent)
		fs.PrintDefaults()
		return nil
	}

	return sub.Execute(fs.Args())
}

func lookup(name string, src []Commander) (Commander, bool) {
	for _, cmd := range src {
		if cmd.equal(name) {
			return cmd, true
		}
	}
	return nil, false
}

func (group *CommandGroup) equal(name string) bool {
	return group.Name == name
}

func (sub *Subcommand) equal(name string) bool {
	return sub.Name == name
}

func (cmd *Command) usage() {
	if cmd.Description != "" {
		fmt.Printf("Description: %s\n\n", cmd.Description)
	}
	fmt.Printf("Usage: %s [version] | <command> ...\n\n", cmd.Name)
	fmt.Print("List of commands:\n")
	for _, sub := range cmd.Subcommands {
		sub.shortPrint()
	}
}

func (group *CommandGroup) usage(parent string) {
	if group.Description != "" {
		fmt.Printf("Description: %s\n\n", group.Description)
	}
	fmt.Printf("Usage: %s %s <command> ...\n\n", parent, group.Name)
	fmt.Print("List of commands:\n")
	for _, sub := range group.Subcommands {
		sub.shortPrint()
	}
}

func (sub *Subcommand) usage(parent string) {
	if sub.Description != "" {
		fmt.Printf("Description: %s\n\n", sub.Description)
	}
	if sub.Flags == nil {
		fmt.Printf("Usage: %s %s [<args>]\n", parent, sub.Name)
		return
	}
	fmt.Printf("Usage: %s %s [<flags>] [<args>]\n\n", parent, sub.Name)
	fmt.Print("List of flags:\n")
}

func (group *CommandGroup) shortPrint() {
	fmt.Printf("\t%s\t%s\n", group.Name, group.Description)
}

func (sub *Subcommand) shortPrint() {
	fmt.Printf("\t%s\t%s\n", sub.Name, sub.Description)
}
