package nginx

import (
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os/exec"
	"strings"
	"syscall"
)

var container string

func Setup(ctx *cli.Context) error {
	if err := SetConfigDir(ctx.String("nginx.conf")); err != nil {
		return fmt.Errorf("set nginx config directory failed.%v", err)
	}
	container = ctx.String("container")
	return nil
}

func checkNginx() error {
	result, err := execNginxCmd("nginx", "-t")
	if strings.Contains(result, "failed") {
		return errors.New(result)
	}
	return err
}
func execNginxCmd(commands ...string) (string, error) {
	args := commands[1:]
	name := commands[0]
	if container != "" {
		args = append([]string{"exec", "-i", container}, commands...)
		name = "docker"
	}
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	output, err := cmd.CombinedOutput()
	log.Printf("%s\n%s", cmd.String(), string(output))
	if err != nil {
		err = errors.New(string(output))
	}
	return string(output), err
}

func Reload() error {
	_, err := execNginxCmd("nginx", "-s", "reload")
	return err
}
