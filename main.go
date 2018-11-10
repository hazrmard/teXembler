package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"text/template"

	"github.com/spf13/viper"
)

func main() {

	var configFpath = flag.String("config", "test/config", `Config file containing
	version information. Without extension.`)

	var configDir = path.Dir(*configFpath)
	var configFile = path.Base(*configFpath)

	viper.SetConfigName(configFile)
	viper.AddConfigPath(configDir)
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}

	settings := viper.AllSettings()

	SiftSettings(settings, "version")

	for _, v := range settings["version"].([]interface{}) {
		version := v.(map[string]interface{})

		tempFiles := make([]string, len(version["files"].([]interface{})))

		for i, f := range version["files"].([]interface{}) {
			f := f.(string)
			fpath := path.Join(configDir, f)
			tempFile, err := ioutil.TempFile(path.Dir(fpath), path.Base(fpath))
			if err != nil {
				panic(err)
			}
			defer os.Remove(tempFile.Name())
			tempFiles[i] = tempFile.Name()

			tmplVer := template.Must(template.New(path.Base(fpath)).ParseFiles(fpath))
			if err := tmplVer.Execute(tempFile, version); err != nil {
				panic(err)
			}
			tempFile.Close()
		}
		version["files"] = tempFiles

		for _, c := range version["cmd"].([]interface{}) {
			cmd := c.([]interface{})
			parsedCmd := make([]string, len(cmd))

			for j, p := range cmd {
				part := p.(string)
				var buff bytes.Buffer

				tmplPart := template.Must(template.New("part").Parse(part))
				tmplPart.Execute(&buff, version)
				parsedCmd[j] = buff.String()
			}

			command := exec.Command(parsedCmd[0], parsedCmd[1:]...)
			command.Stdout = os.Stdout
			if err := command.Run(); err != nil {
				fmt.Println(parsedCmd, "- err:", err)
			}
		}
	}
}
