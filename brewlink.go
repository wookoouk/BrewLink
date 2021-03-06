package main

import (
	"encoding/json"
	"errors"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/kardianos/osext"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"path/filepath"
	"fmt"
)

var (
	config Config
	pathToMe string
	err error
)

type Config struct {
	CellarPath   string `json:"CellarPath"`
	SoftwarePath string `json:"SoftwarePath"`
}

func main() {

	//pathToMe, err = filepath.Abs(filepath.Dir(os.Args[0]))
	pathToMe, err = osext.ExecutableFolder()
	check(err)

	err := loadConfig()
	check(err)

	//new cli app
	app := cli.NewApp()
	app.Version = "0.4.2"
	app.Name = "BrewLink"
	app.Usage = "Link software installed with brew to somewhere else"
	app.Commands = []cli.Command{
		{
			Name:"link",
			Aliases:[]string{"l"},
			Usage: "Link package",
			Action: func(c *cli.Context) {
				if len(c.Args()) == 1 {
					magic(c.Args().First(), *c)
					//println("Hello", c.Args()[0])
				} else {
					//show user the way to use the app
					cli.ShowAppHelp(c)
				}
			},
		},
		{
			Name:"show",
			Aliases:[]string{"s"},
			Usage:"Show status of all packages",
			Action: func(c *cli.Context) {

				showStatus()

				//if (len(c.Args()) > 0) {
				//	SetMemory(c.Args().First())
				//	PrintSuccess(SetMemoryMessage, GetMemory())
				//} else {
				//	PrintSuccess(GetMemoryMessage, GetMemory())
				//}
			},
		},
	}

	//process the above
	app.Run(os.Args)
}
func listNameVersion(dir string) []string {
	found := []string{}
	folders, _ := ioutil.ReadDir(dir)
	for _, f := range folders {
		insidePath := path.Join(dir, f.Name())
		foldersSub, _ := ioutil.ReadDir(insidePath)
		for _, fs := range foldersSub {

			if (dir == config.SoftwarePath) {
				versionFolder := path.Join(insidePath, fs.Name(), "x86_64")
				vExists, vError := exists(versionFolder);
				if (vError == nil && vExists) {
					ss, err := filepath.EvalSymlinks(versionFolder)
					if (err == nil) {
						split := strings.Split(ss, config.CellarPath)
						splitLen := len(split)
						if (splitLen == 2) {
							found = append(found, split[1])
						}
					} else {
						//its not symlinked
					}
				} else {
					//doesnt exists
				}
			} else if (dir == config.CellarPath) {
				versionFolder := path.Join(insidePath, fs.Name())
				split := strings.Split(versionFolder, config.CellarPath)
				splitLen := len(split)
				if (splitLen == 2) {
					found = append(found, split[1])
				}
			}
		}
	}
	tidy := []string{}
	for _, f := range found {
		split := strings.Split(f[1:len(f)], "/")
		tidy = append(tidy, split[0] + "-" + split[1])
	}
	return tidy
}

func installedList() []string {
	return listNameVersion(config.CellarPath)
}
func linkedList() []string {
	return listNameVersion(config.SoftwarePath)
}

func showStatus() {
	installed := installedList()
	linked := linkedList()
	for _, i := range installed {
		found := false
		for _, l := range linked {
			if (i == l) {
				found = true
			}
		}
		if (found) {
			//println(i, "linked")
			PrintRed(i)
		} else {
			PrintGreen(i)
			//println(i, "un-linked")
		}
	}
}

func PrintGreen(s ...interface{}) {
	color.Set(color.FgRed)
	fmt.Fprintln(os.Stderr, s...)
	color.Unset()
}

func PrintRed(s ...interface{}) {
	color.Set(color.FgGreen)
	fmt.Fprintln(os.Stdout, s...)
	color.Unset()
}

func loadConfig() error {

	//path to configFile
	configPath := path.Join(pathToMe, ".brewlink.json")

	//println("looking for config in", configPath)

	//read config
	dat, err := ioutil.ReadFile(configPath)
	check(err)

	//create empty Config struct
	config = Config{}

	//unmarshal config file to Config struct
	err = json.Unmarshal(dat, &config)
	check(err)

	//both paths should exists to begin with
	sExists, sError := exists(config.SoftwarePath)
	check(sError)
	cExists, cError := exists(config.CellarPath)
	check(cError)

	if !sExists {
		return errors.New("it seems that the SoftwarePath in your config does not exist, please check .brewlink.json")
	}
	if !cExists {
		return errors.New("it seems that the CellarPath in your config does not exist, please check .brewlink.json")

	}

	return nil

}

//function taken from http://stackoverflow.com/questions/10510691/how-to-check-whether-a-file-or-directory-denoted-by-a-path-exists-in-golang
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func magic(a string, c cli.Context) {

	//split into chunks at '-'
	splitString := strings.Split(a, "-")

	if len(splitString) < 2 {
		//too short
		println("Error: There should only be one arugument")
		cli.ShowAppHelp(&c)
	} else if len(splitString) > 2 {
		//too long
		println("Error: There should only be one arugument")
		cli.ShowAppHelp(&c)
	} else {
		//just right
		toolName := splitString[0]
		toolVersion := splitString[1]

		//build old path
		oldPath := path.Join(config.CellarPath, toolName, toolVersion)

		//folder above x86_64
		newPathPreX86 := path.Join(config.SoftwarePath, toolName, toolVersion)

		//make link parent folder (mkdir -p)
		err := os.MkdirAll(newPathPreX86, 0755)
		check(err)

		//path to target
		symLinkTarget := path.Join(config.SoftwarePath, toolName, toolVersion, "x86_64")

		//create sym link
		err = os.Symlink(oldPath, symLinkTarget)
		check(err)

		println("The link has been created.", symLinkTarget)

		//println("linking %v to %v", oldPath, symLinkTarget)
	}
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
