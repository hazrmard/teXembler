package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"text/template"

	"github.com/spf13/viper"
)

func main() {

	var configFpath = flag.String("config", "test/config", `Config file containing
	version information. Without extension.`)
	flag.Parse()

	var configDir = path.Dir(*configFpath)
	var configFile = path.Base(*configFpath)

	viper.SetConfigName(configFile)
	viper.AddConfigPath(configDir)
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}

	settings := viper.AllSettings()
	settings["root"] = configDir

	SiftSettings(settings, "version")

	// Generate each version sequentially
	for _, v := range settings["version"].([]interface{}) {
		version := v.(map[string]interface{})
		flist := version["files"].([]interface{}) // reference to original file names
		tempFiles := make([]string, len(flist))   // reference to file names accessible by versions after templating
		fObjs := make([]*os.File, len(flist))     // reference to *File objects

		// For each file, create a temporary file to output template results,
		// overwrite file names with temporary file locations and preserve
		// originals.
		for i, f := range flist {
			f := f.(string)
			fpath := path.Join(configDir, f)
			tempFile, err := ioutil.TempFile(path.Dir(fpath), "_temp"+path.Base(fpath))
			if err != nil {
				panic(err)
			}
			defer os.Remove(tempFile.Name())
			fObjs[i] = tempFile
			if relPath, err := filepath.Rel(configDir, tempFile.Name()); err != nil {
				panic(err)
			} else {
				tempFiles[i] = filepath.ToSlash(relPath)
			}
		}
		version["files"] = tempFiles

		// Now that temporary file handles have been generated, create templates
		// with references to temporary file names.
		for i, f := range flist {
			f := f.(string)
			fpath := path.Join(configDir, f)
			tmplVer := template.Must(template.New(path.Base(fpath)).ParseFiles(fpath))
			if err := tmplVer.Execute(fObjs[i], version); err != nil {
				panic(err)
			}
			fObjs[i].Close()
		}

		// After processing templates, run all commands for that version.
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
			command.Stderr = os.Stderr
			if err := command.Run(); err != nil {
				fmt.Println(parsedCmd, "- err:", err)
			}
		}
	}
}
