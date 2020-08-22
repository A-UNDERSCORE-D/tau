package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/anmitsu/go-shlex"
	"github.com/spf13/pflag"
)

var (
	configPath string
	command    string
	verbose    bool
)

func main() {
	pflag.StringVarP(&configPath, "config", "c", os.ExpandEnv("$HOME/.config/tau.toml"), "sets the config file to use")
	pflag.BoolVarP(&verbose, "verbose", "v", false, "enables verbose logging")
	pflag.StringVar(&command, "command", "", "sets the command to use when uploading")
	pflag.Args()

	pflag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fmt.Printf("%s [FLAGS]... [FILE]...\n", os.Args[0])
		pflag.PrintDefaults()
	}

	pflag.Parse()

	args := pflag.Args()

	conf, err := getConf(configPath)
	if err != nil {
		fmt.Printf("could not get config file: %s\n", err)
		return
	}
	for _, targetFile := range args {
		newName, err := doTransform(targetFile, conf)
		if err != nil {
			fmt.Printf("could not transform path: %s\n", err)
			continue
		}

		if err := Execute(command, targetFile, newName); err != nil {
			fmt.Printf("error occurred while executing command: %s\n", err)
			continue
		}
	}

}

func verboseLogf(format string, args ...interface{}) {
	if !verbose {
		return
	}
	fmt.Printf(format+"\n", args...)
}

func doTransform(path string, conf config) (string, error) {

	trimmedPath := filepath.Base(path)
	verboseLogf("%q trimmed to %q", path, trimmedPath)

	for name, c := range conf {

		subMatches := c.compiledMatcher.FindStringSubmatchIndex(trimmedPath)
		if subMatches == nil {
			continue
		}

		out := c.compiledMatcher.ExpandString(nil, c.Transform, trimmedPath, subMatches)
		fmt.Printf("matched %q against matcher %q. New name: %q\n", trimmedPath, name, string(out))
		return string(out), nil
	}

	return "", fmt.Errorf("no matches found for %q", path)
}

func printCmd(cmd []string) string {
	out := strings.Builder{}

	for _, v := range cmd {
		if strings.Contains(v, " ") {
			out.WriteString(fmt.Sprintf("%q", v))
		} else {
			out.WriteString(v)
		}
		out.WriteRune(' ')
	}

	return out.String()
}

func Execute(command, targetFile, newName string) error {
	if command == "" {
		fmt.Println("No command supplied, not executing")
		return nil
	}

	split, err := shlex.Split(command, true)
	if err != nil {
		return fmt.Errorf("could not parse command: %w", err)
	}

	replacer := strings.NewReplacer("$target", newName, "$source", targetFile)

	for i, l := range split[1:] {
		split[i+1] = replacer.Replace(l)
	}

	fmt.Println("executing: ", printCmd(split))

	cmd := exec.Command(split[0], split[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
