package gophkeeper

import (
	"errors"
	"flag"

	"github.com/sergeizaitcev/gophkeeper/pkg/cli"
	"github.com/sergeizaitcev/gophkeeper/version"
)

// errArgsTooSmall возвращается, когда передано слишком мало аргументов.
var errArgsTooSmall = errors.New("there are too few arguments")

// Command определяет консольное приложение gophkeeper.
var Command = &cli.Command{
	Name:        "gk",
	Description: "gophkeeper client",
	Version:     version.Version,
	Subcommands: []cli.Commander{
		&cli.CommandGroup{
			Name:        "remote",
			Description: "remote server settings",
			Subcommands: []cli.Commander{
				&cli.Subcommand{
					Name:        "set",
					Description: "setting the address of the remote server",
					Execute:     RemoteSet,
				},
				&cli.Subcommand{
					Name:        "show",
					Description: "shows the address of the remote server",
					Execute:     RemoteShow,
				},
			},
		},
		&cli.Subcommand{
			Name:        "login",
			Description: "authorization on a remote server",
			Flags: func(fs *flag.FlagSet) {
				fs.StringVar(&flagUsername, "u", "", "user's login")
				fs.StringVar(&flagPassword, "p", "", "user's password")
			},
			Execute: Login,
		},
		&cli.CommandGroup{
			Name:        "add",
			Description: "adding new data with encryption to the vault",
			Subcommands: []cli.Commander{
				&cli.Subcommand{
					Name:        "card",
					Description: "adding a bank card number to the vault",
					Flags: func(fs *flag.FlagSet) {
						fs.StringVar(&flagDescription, "d", "", "description of the data")
					},
					Execute: AddBankCard,
				},
				&cli.Subcommand{
					Name:        "logpass",
					Description: "adding a username-password to the vault",
					Flags: func(fs *flag.FlagSet) {
						fs.StringVar(&flagDescription, "d", "", "description of the data")
					},
					Execute: AddUsernamePassword,
				},
				&cli.Subcommand{
					Name:        "file",
					Description: "adding a file to the vault",
					Flags: func(fs *flag.FlagSet) {
						fs.StringVar(&flagDescription, "d", "", "description of the data")
					},
					Execute: AddFile,
				},
			},
		},
		&cli.Subcommand{
			Name:        "rm",
			Description: "deleting data from the vault",
			Execute:     Remove,
		},
		&cli.Subcommand{
			Name:        "sync",
			Description: "synchronizing files with a remote server",
			Execute:     Sync,
		},
		&cli.Subcommand{
			Name:        "show",
			Description: "show data in the vault",
			Flags: func(fs *flag.FlagSet) {
				fs.StringVar(&flagOutput, "o", "", "path to output")
			},
			Execute: Show,
		},
		&cli.Subcommand{
			Name:        "ls",
			Description: "show a list of all data in the vault",
			Execute:     List,
		},
	},
}
