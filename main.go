package main

import (
	"github.com/urfave/cli/v2"
	"log"
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
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "dns",
				Usage:   "Name of dns provider",
				EnvVars: []string{"DNS_PROVIDER"},
			},
			&cli.StringFlag{
				Name:    "log",
				Usage:   "log directory",
				EnvVars: []string{"NWM_LOG"},
			},
			&cli.StringFlag{
				Name:    "email",
				Usage:   "Email used for cert registration and recovery contact",
				EnvVars: []string{"EMAIL"},
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
			&cli.StringFlag{
				Name:    "path",
				Value:   ".",
				Usage:   "Directory to use for storing the data.",
				EnvVars: []string{"CERT_PATH"},
			},
			&cli.StringFlag{
				Name:    "container",
				Usage:   "if your nginx is running in docker mode,set the nginx container name to this value",
				EnvVars: []string{"CONTAINER"},
			},
			&cli.StringFlag{
				Name:    "cron",
				Usage:   "cron expression for renew check (default: 0 0 0 * * ?) ",
				EnvVars: []string{"CRON"},
			},
			&cli.StringFlag{
				Name:    "url",
				Usage:   "CA hostname (and optionally :port). The server certificate must be trusted in order to avoid further modifications to the client.(default: https://acme-v02.api.letsencrypt.org/directory) ",
				EnvVars: []string{"CA_URL"},
			},
		},
	}
}
