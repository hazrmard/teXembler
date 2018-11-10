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

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	settings := viper.AllSettings()
	files := viper.GetStringSlice("files")
	versions := settings["version"].([]interface{})
	cmds := viper.Get("cmd").([]interface{})

	for _, v := range versions {
		version := v.(map[string]interface{})

		if version["files"] == nil {
			version["origFiles"] = files
		} else {
			version["origFiles"] = version["files"]
		}
		if version["cmd"] == nil {
			version["cmd"] = cmds
		}

		tempFiles := make([]string, len(version["origFiles"].([]string)))

		for i, f := range version["origFiles"].([]string) {
			fpath := path.Join(configDir, f)
			tempFile, err := ioutil.TempFile(path.Dir(fpath), path.Base(fpath))
			defer os.Remove(tempFile.Name())
			tempFiles[i] = tempFile.Name()

			tmplVer := template.Must(template.New(path.Base(fpath)).ParseFiles(fpath))
			err = tmplVer.Execute(tempFile, version)
			tempFile.Close()
			if err != nil {
				panic(err)
			}
		}
		version["files"] = tempFiles

		for _, c := range cmds {
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
