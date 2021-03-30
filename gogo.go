package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/fatih/color"
	"github.com/goccy/go-yaml"
)

const gogoYaml = ".gogo.yaml"
const gogoPasswd = ".gogopasswd"

func usage() {
	fmt.Println(`
	NAME:
	   gogo - a tool to run a command in a given environment
	
	USAGE:
		gogo <env> command [command arguments...]
		gogo [options] <env>
	
	OPTIONS:
	   --list     list variables
	   --encrypt  encrypt .gogo.yaml
	   --decrypt  decrypt .gogo.yaml
	   --version  display version
	   --dry      dry mode (default: false)
	   --verbose  verbose mode (default: false)
	   --help     show help (default: false)
		`)
}

func getConfigFilePath() string {
	currentPath, err := os.Getwd()
	failIf(err, "Could not get current folder")
	currentGogoYaml := currentPath + "/" + gogoYaml
	if _, errStat := os.Stat(currentGogoYaml); errStat == nil {
		return currentGogoYaml
	}

	currentUser, errCurrent := user.Current()
	failIf(errCurrent, "Could not get current user")

	homeGogoYaml := currentUser.HomeDir + "/" + gogoYaml
	if _, errStat := os.Stat(homeGogoYaml); errStat == nil {
		return homeGogoYaml
	}

	fmt.Printf("could not find any %v", gogoYaml)
	return ""
}

func getPasswordFilePath() string {
	usr, err := user.Current()
	failIf(err, "Could not get current user")
	return usr.HomeDir + "/" + gogoPasswd
}

// return a 32 bites string
func getPassword() string {
	gogoPasswdContent, err := ioutil.ReadFile(getPasswordFilePath())
	failIf(err, "could not read "+gogoPasswd)

	return string(gogoPasswdContent)
}

func savePassword(input string) {
	key32 := rpad(input, "x", 32)
	bytes := []byte(key32)

	gogopasswdFile := getPasswordFilePath()
	if _, err := os.Stat(gogopasswdFile); err == nil {
		fmt.Printf("~/%v already exists", gogoPasswd)
		return
	}

	ioutil.WriteFile(gogopasswdFile, bytes, 0644)
}

type Configugration struct {
	Encrypted bool                         `yaml:"encrypted"`
	Envs      map[string]map[string]string `yaml:"envs"`
}

func getAllConfig() Configugration {
	yamlFile, err := ioutil.ReadFile(getConfigFilePath())
	failIf(err, "could not read "+gogoYaml)

	var config Configugration

	err = yaml.Unmarshal(yamlFile, &config)
	failIf(err, "could not unmarshal "+gogoYaml)

	return config
}

func getConfig(environment string) (map[string]string, error) {
	if len(environment) == 0 {
		return nil, errors.New("Environment required")
	}
	allConfig := getAllConfig()
	config := allConfig.Envs[environment]
	if allConfig.Encrypted {
		encryptionKey := getPassword()
		for varName, varValue := range config {
			config[varName] = decryptString(varValue, encryptionKey)
		}
	}

	return config, nil
}

func saveConfig(config Configugration) {
	bytes, err := yaml.Marshal(config)
	failIf(err, "could not marshal config")

	ioutil.WriteFile(getConfigFilePath(), bytes, 0644)
}

func addToConfig(environment string, key string, value string) error {
	config := getAllConfig()

	if config.Encrypted {
		encryptionKey := getPassword()
		config.Envs[environment][key] = encryptString(value, encryptionKey)
	} else {
		config.Envs[environment][key] = value
	}

	bytes, err := yaml.Marshal(config)
	failIf(err, "could not marshal "+gogoYaml)

	ioutil.WriteFile(getConfigFilePath(), bytes, 0)

	return nil
}

func main() {

	var list bool
	var add bool
	var encrypt bool
	var decrypt bool
	var version bool
	var dry bool
	var verbose bool

	f := flag.NewFlagSet("flag", flag.ContinueOnError)

	f.BoolVar(&list, "list", false, "list vars in a given envs")
	f.BoolVar(&add, "add", false, "add a var in a given env")
	f.BoolVar(&encrypt, "encrypt", false, "encrypt "+gogoYaml)
	f.BoolVar(&decrypt, "decrypt", false, "decrypt "+gogoYaml)
	f.BoolVar(&version, "version", false, "print version")
	f.BoolVar(&dry, "dry", false, "dry mode")
	f.BoolVar(&verbose, "verbose", false, "verbose mode")

	indexEnv := -1
	for i := range os.Args {
		if i > 0 && !strings.HasPrefix(os.Args[i], "-") {
			indexEnv = i
			break
		}
	}

	args := os.Args[1:]
	color := color.New(color.FgHiBlack).SprintFunc()
	var environment string

	if indexEnv > 0 {
		args = os.Args[1:indexEnv]
		environment = os.Args[indexEnv]
	}

	err := f.Parse(args)
	if err != nil {
		fmt.Println("err %w", err)
		//usage()
		return
	}

	if list {
		if len(environment) == 0 {
			environment = ask("Enter env name: ")
		}
		config, _ := getConfig(environment)

		for varName, varValue := range config {
			fmt.Println(color(" - " + varName + "=" + varValue))
		}
	} else if add {
		if len(environment) == 0 {
			environment = ask("Enter env name: ")
		}
		varName := ask("Enter variable name: ")
		varValue := ask("Enter variable value: ")

		addToConfig(environment, varName, varValue)
		fmt.Println(color(fmt.Sprintf("Var added to env %s: %s=%s", environment, varName, varValue)))

	} else if version {
		fmt.Println("gogo", Version)

	} else if encrypt {
		encryptionKey := getPassword()

		if len(encryptionKey) == 0 {
			input := ask("Enter password: ")
			savePassword(input)
			encryptionKey = getPassword()
		}

		config := getAllConfig()
		if config.Encrypted {
			fmt.Println("Already encrypted")
			return
		}

		config.Encrypted = true
		for environment, vars := range config.Envs {
			for varName, varValue := range vars {
				config.Envs[environment][varName] = encryptString(varValue, encryptionKey)
			}
		}

		saveConfig(config)
		fmt.Printf("%s", "Encryption done ✅")

	} else if decrypt {
		encryptionKey := getPassword()

		if len(encryptionKey) == 0 {
			input := ask("Enter password: ")
			savePassword(input)
			encryptionKey = getPassword()
		}

		config := getAllConfig()
		if config.Encrypted == false {
			fmt.Println("Already decrypted")
			return
		}

		config.Encrypted = false
		for environment, vars := range config.Envs {
			for varName, varValue := range vars {
				config.Envs[environment][varName] = decryptString(varValue, encryptionKey)
			}
		}

		saveConfig(config)
		fmt.Println("Encryption done ✅")

	} else if environment != "" {
		// run command
		if verbose {
			fmt.Println(color("env: " + environment))
		}
		config, _ := getConfig(environment)

		cmdArray := os.Args[indexEnv+1:]

		for key, name := range config {
			if verbose {
				fmt.Println(color(" - inject " + key + "=" + name))
			}
			for i := 0; i < len(cmdArray); i++ {
				cmdArray[i] = strings.ReplaceAll(cmdArray[i], "$"+key, name)
			}
		}

		firstArg := strings.Split(cmdArray[0], " ")
		command := firstArg[0]
		rest := firstArg[1:]
		arguments := append(rest, cmdArray[1:]...)

		cmd := exec.Command(command, arguments...)
		if verbose {
			fmt.Println(color(cmd))
		}

		cmd.Env = os.Environ()

		for key, name := range config {
			cmd.Env = append(cmd.Env, key+"="+name)
		}

		stdout, _ := cmd.StdoutPipe()
		cmd.Start()

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			m := scanner.Text()
			fmt.Println(m)
		}

		cmd.Wait()
	} else {
		usage()
	}

}
