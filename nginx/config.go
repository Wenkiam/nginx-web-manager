package nginx

import (
	"errors"
	"io"
	"log"
	"os"
	"strings"
)

type Config struct {
	Name    string `json:"name"`
	Content string `json:"content"`
	Reload  bool   `json:"reload"`
}

var (
	configPath = "/etc/nginx/conf.d/"
)

func AllConfigs() ([]Config, error) {
	dirs, err := os.ReadDir(configPath)
	if err != nil {
		return nil, err
	}
	var configs []Config
	for _, dir := range dirs {
		if !strings.HasSuffix(dir.Name(), ".conf") {
			continue
		}
		if dir.IsDir() {
			continue
		}
		file, err := os.Open(configPath + dir.Name())
		if err != nil {
			continue
		}
		bytes, err := io.ReadAll(file)
		if err != nil {
			continue
		}
		configs = append(configs, Config{
			Name:    dir.Name(),
			Content: string(bytes),
		})
	}
	return configs, nil
}

func SaveConfig(config *Config) error {
	if config.Name == "" {
		return errors.New("file name is empty")
	}
	if config.Content == "" {
		return errors.New("content is empty")
	}
	conf := config.Name
	if !strings.HasSuffix(conf, ".conf") {
		conf = conf + ".conf"
	}
	conf = configPath + conf
	oldContent, err := os.ReadFile(conf)
	if os.IsNotExist(err) {
		_, err = os.Create(conf)
	}
	if err != nil {
		log.Printf("read or create config file %s failed", conf)
		return err
	}

	err = os.WriteFile(conf, []byte(config.Content), 0644)
	if err != nil {
		log.Printf("write content to %s failed.reason:%v ", conf, err)
		return err
	}
	if !config.Reload {
		return nil
	}
	err = checkNginx()
	if err != nil {
		if os.WriteFile(conf, oldContent, 0644) != nil {
			log.Printf("roll back config content failed")
		}
		return err
	}
	return Reload()
}
func SetConfigDir(dir string) error {
	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}
	_, err := os.Stat(dir)
	if err != nil {
		return err
	}
	configPath = dir
	return err
}

func GetPath() string {
	return configPath
}
func DelConf(conf string) error {
	if !strings.HasSuffix(conf, ".conf") {
		conf = conf + ".conf"
	}
	err := os.Remove(configPath + conf)
	if err != nil {
		return err
	}
	return Reload()
}
