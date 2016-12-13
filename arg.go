package main

import (
	"os"
	"sort"

	"github.com/urfave/cli"
)

func arg() *cli.App {
	app := cli.NewApp()
	app.Name = `wstail`
	app.Usage = `convert "tail -f" as websocket output`
	app.Version = version

	app.UsageText = app.Name + ` [OPTION]...`
	app.Authors = []cli.Author{
		cli.Author{
			Name:  `Zheng Kai`,
			Email: `zhengkai@gmail.com`,
		},
	}
	app.HideHelp = true

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "Load configuration from `FILE`",
			Value: configFileName,
		},
		cli.StringFlag{
			Name:  "whitelist, w",
			Usage: "Load whitelist from `FILE`",
			Value: whitelistFileName,
		},
		cli.StringFlag{
			Name:  "listen, l",
			Usage: "WebSocket HTTP listen `ADDRESS`",
			Value: httpListen,
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))

	app.Flags = append(
		app.Flags,
		cli.BoolFlag{
			Name:  "help, h",
			Usage: "show this help message and exit",
		},
	)

	app.Run(os.Args)

	set := flag.NewFlagSet("contrive", 0)

	nc := cli.NewContext(app, set, c)

	return arg
}
