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

	"github.com/denisbrodbeck/machineid"
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
		--init    create .gogo.yaml
		--list    list variables
		--encrypt  encrypt .gogo.yaml
		--decrypt  decrypt .gogo.yaml
		--version  display version
		--dry      dry mode (default: false)
		--verbose  verbose mode (default: false)
		--help     show help (default: false)`)
}

func getCurrentDirConfigFilePath() string {
	currentPath, err := os.Getwd()
	failIf(err, "Could not get current folder")
	return currentPath + "/" + gogoYaml
}

func getConfigFilePath() string {
	currentGogoYaml := getCurrentDirConfigFilePath()
	if _, errStat := os.Stat(currentGogoYaml); errStat == nil {
		return currentGogoYaml
	}

	currentUser, errCurrent := user.Current()
	failIf(errCurrent, "Could not get current user")

	homeGogoYaml := currentUser.HomeDir + "/" + gogoYaml
	if _, errStat := os.Stat(homeGogoYaml); errStat == nil {
		return homeGogoYaml
	}

	fmt.Printf("could not find any %v\n", gogoYaml)
	return ""
}

/*
 * returns a 32 string used to encrypt user's password
 */
func getMachineID() string {
	uuid64, err := machineid.ProtectedID("gogo")
	failIf(err, "could not read machine id")

	return uuid64[:32]
}

func getPasswordFilePath() string {
	usr, err := user.Current()
	failIf(err, "Could not get current user")
	return usr.HomeDir + "/" + gogoPasswd
}

// return a 32 bites string
func getPassword() string {
	if _, err := os.Stat(getPasswordFilePath()); os.IsNotExist(err) {
		return ""
	}

	gogoPasswdContent, err := ioutil.ReadFile(getPasswordFilePath())
	failIf(err, "could not read "+gogoPasswd)

	password := decryptString(string(gogoPasswdContent), getMachineID())

	return password
}

func savePassword(password string) {
	password32 := rpad(password, "x", 32)
	encryptedPassword32 := encryptString(string(password32), getMachineID())
	bytes := []byte(encryptedPassword32)

	gogopasswdFile := getPasswordFilePath()
	if _, err := os.Stat(gogopasswdFile); err == nil {
		fmt.Printf("~/%v already exists\n", gogoPasswd)
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

	if config.Envs[environment] == nil {
		config.Envs[environment] = map[string]string{}
	}

	if config.Encrypted {
		encryptionKey := getPassword()
		config.Envs[environment][key] = encryptString(value, encryptionKey)
	} else {
		config.Envs[environment][key] = value
	}

	bytes, err := yaml.Marshal(config)
	failIf(err, "could not marshal "+gogoYaml)

	ioutil.WriteFile(getConfigFilePath(), bytes, 0644)

	return nil
}

func main() {

	var list bool
	var init bool
	var add bool
	var encrypt bool
	var decrypt bool
	var version bool
	var dry bool
	var verbose bool

	f := flag.NewFlagSet("flag", flag.ContinueOnError)

	f.BoolVar(&init, "init", false, "create .gogo.yaml")
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
	colorInfo := color.New(color.FgHiBlack).SprintFunc()
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
			fmt.Println(colorInfo(" - " + varName + "=" + varValue))
		}
	} else if init {
		if len(environment) == 0 {
			environment = ask("Enter env name: ")
		}
		currentGogoYaml := getCurrentDirConfigFilePath()
		if _, errStat := os.Stat(currentGogoYaml); errStat == nil {
			fmt.Printf("~/%v already exists\n", currentGogoYaml)
			return
		}

		bootstrapConfig := fmt.Sprintf("encrypted: false\nenvs:\n    %v:\n", environment)

		ioutil.WriteFile(currentGogoYaml, []byte(bootstrapConfig), 0644)

		fmt.Println(colorInfo(fmt.Sprintf(".gogo.yaml created with env %s\n", environment)))

	} else if add {
		if len(environment) == 0 {
			environment = ask("Enter env name: ")
		}
		varName := ask("Enter variable name: ")
		varValue := ask("Enter variable value: ")

		addToConfig(environment, varName, varValue)
		fmt.Println(colorInfo(fmt.Sprintf("Var added to env %s: %s=%s\n", environment, varName, varValue)))

	} else if version {
		fmt.Println(Version)

	} else if encrypt {
		encryptionKey := getPassword()

		if len(encryptionKey) == 0 {
			input := ask("Enter password (max. 32): ")
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
		fmt.Println("Encryption done ???")

	} else if decrypt {
		config := getAllConfig()
		if config.Encrypted == false {
			fmt.Println("Already decrypted")
			return
		}

		encryptionKey := getPassword()

		if len(encryptionKey) == 0 {
			input := ask("Enter password: ")
			savePassword(input)
			encryptionKey = getPassword()
		}

		config.Encrypted = false
		for environment, vars := range config.Envs {
			for varName, varValue := range vars {
				config.Envs[environment][varName] = decryptString(varValue, encryptionKey)
			}
		}

		saveConfig(config)
		fmt.Println("Decryption done ???")

	} else if environment != "" {
		// run command
		if verbose {
			fmt.Println(colorInfo("env: " + environment))
		}
		config, _ := getConfig(environment)

		cmdArray := os.Args[indexEnv+1:]

		for key, name := range config {
			if verbose {
				fmt.Println(colorInfo(" - inject " + key + "=" + name))
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
			fmt.Println(colorInfo(cmd))
		}

		cmd.Env = os.Environ()

		for key, name := range config {
			cmd.Env = append(cmd.Env, key+"="+name)
		}

		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()
		cmd.Start()

		scannerOut := bufio.NewScanner(stdout)
		for scannerOut.Scan() {
			m := scannerOut.Text()
			fmt.Println(m)
		}

		scannerErr := bufio.NewScanner(stderr)
		colorError := color.New(color.FgRed).SprintFunc()
		for scannerErr.Scan() {
			m := scannerErr.Text()
			fmt.Println(colorError(m))
		}

		cmd.Wait()
	} else {
		usage()
	}

}
