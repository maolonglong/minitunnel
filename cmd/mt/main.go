package main

import (
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v2"

	"go.chensl.me/minitunnel/internal/tunnel"
)

func main() {
	log.SetTimeFormat(time.DateTime)

	app := &cli.App{
		Name:  "mt",
		Usage: "simple CLI tool for making tunnels to localhost",
		Commands: []*cli.Command{
			{
				Name:      "local",
				Usage:     "starts a local proxy to the remote server",
				ArgsUsage: "<local_port>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "local-host",
						Aliases: []string{"l"},
						Value:   "localhost",
						Usage:   "the local host to expose",
					},
					&cli.StringFlag{
						Name:    "to",
						Aliases: []string{"t"},
						EnvVars: []string{"MT_SERVER"},
						Value:   "minitunnel.icu",
						Usage:   "address of the remote server to expose local ports to",
					},
				},
				Action: func(cCtx *cli.Context) error {
					if cCtx.NArg() != 1 {
						cli.ShowSubcommandHelpAndExit(cCtx, 2)
					}
					localPort, err := strconv.ParseInt(cCtx.Args().Get(0), 10, 64)
					if err != nil {
						return err
					}
					cli, err := tunnel.NewClient(
						cCtx.String("to"),
						cCtx.String("local-host"),
						int(localPort),
					)
					if err != nil {
						return err
					}
					return cli.Run()
				},
			},
			{
				Name:  "server",
				Usage: "runs the remote proxy server.",
				Action: func(cCtx *cli.Context) error {
					if cCtx.NArg() != 0 {
						cli.ShowSubcommandHelpAndExit(cCtx, 2)
					}
					return tunnel.NewServer().Run()
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Error("run app failed", "err", err)
	}
}
