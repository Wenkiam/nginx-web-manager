package main

import (
	"github.com/urfave/cli/v2"
	"log"
	"nwm/acme"
	"nwm/nginx"
	"nwm/web"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "nwm"
	app.HelpName = app.Name
	app.Usage = "A web server to manage nginx and certificates"
	app.Commands = createCommands()
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("start app failed.%v", err)
	}
}

func createCommands() []*cli.Command {
	return cli.Commands{
		startCmd(),
	}
}

func startCmd() *cli.Command {
	return &cli.Command{
		Name:  "start",
		Usage: "Start a web server to manage nginx and certificates",
		Before: func(ctx *cli.Context) error {
			err := setupLog(ctx)
			if err != nil {
				return err
			}
			err = nginx.Setup(ctx)
			if err != nil {
				return err
			}
			web.Setup(ctx)
			return nil
		},
		Action: func(ctx *cli.Context) error {
			return web.StartServer()
		},
		Flags: createFlags(),
	}
}
func createFlags() []cli.Flag {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:    "log",
			Usage:   "log directory",
			EnvVars: []string{"NWM_LOG"},
		},
		&cli.IntFlag{
			Name:    "port",
			Value:   8080,
			Usage:   "port of web server",
			Aliases: []string{"p"},
			EnvVars: []string{"PORT"},
		},
		&cli.StringFlag{
			Name:    "nginx.conf",
			Value:   "/etc/nginx/conf.d/",
			Usage:   "Directory of nginx config files",
			EnvVars: []string{"NGINX_CONF"},
		},
	}
	flags = append(flags, web.Flags...)
	flags = append(flags, acme.Flags...)
	return flags
}
