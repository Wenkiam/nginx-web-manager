package acme

import "github.com/urfave/cli/v2"

var Flags = []cli.Flag{
	&cli.StringFlag{
		Name:    "email",
		Usage:   "Email used for cert registration and recovery contact",
		EnvVars: []string{"EMAIL"},
	},
	&cli.StringFlag{
		Name:    "dns",
		Usage:   "Name of dns provider",
		EnvVars: []string{"DNS_PROVIDER"},
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
}
