package cmd

import (
	"fmt"
	"github.com/HARCHHI/gosible/task"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type device struct {
	User     string
	Addr     string
	Port     string
	Password string
}

// gosibleConfig struct for config yaml
type gosibleConfig struct {
	Configs *struct {
		PrivateKey string `yaml:"privateKey"`
		Proxy      *device
	}
	Devices []*device
	Copy    []*task.CopyInfo
	Execute []string
}

func getConnInfo(taskConfig *gosibleConfig) []*task.ConnInfo {
	var proxy *task.ConnInfo
	result := []*task.ConnInfo{}
	privateKey, err := ioutil.ReadFile(taskConfig.Configs.PrivateKey)

	if err != nil {
		log.Printf("%+v", err)
	}
	if proxyConfig := taskConfig.Configs.Proxy; proxyConfig != nil {
		proxy = &task.ConnInfo{
			Addr:       fmt.Sprintf("%s:%s", proxyConfig.Addr, proxyConfig.Port),
			User:       proxyConfig.User,
			Password:   proxyConfig.Password,
			PrivateKey: privateKey,
		}
	}
	for _, dev := range taskConfig.Devices {
		result = append(result, &task.ConnInfo{
			Addr:       fmt.Sprintf("%s:%s", dev.Addr, dev.Port),
			User:       dev.User,
			Password:   dev.Password,
			PrivateKey: privateKey,
			Proxy:      proxy,
		})
	}

	return result
}

func parseYaml(filePath string) *gosibleConfig {
	rawYaml, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	config := &gosibleConfig{}
	err = yaml.Unmarshal(rawYaml, config)
	if err != nil {
		log.Fatal(err)
	}

	return config
}

// Execute execute command
func Execute() error {
	var configFile string
	rootCmd := &cobra.Command{
		Use:   "gosible",
		Short: "simple tool for automate app",
		Run: func(cmd *cobra.Command, args []string) {
			configs := parseYaml(configFile)
			connInfos := getConnInfo(configs)
			manager := task.NewManager(connInfos, configs.Copy, configs.Execute)
			go manager.Start(1)
			for execLog := range manager.LogChan() {
				log.Printf("----%+v----\n%+v", execLog.Device, execLog.Log)
			}
		},
	}
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "f", "", "config file, accept yaml")

	return rootCmd.Execute()
}
