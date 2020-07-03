package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/anmitsu/go-shlex"
)

var (
	configPath string
	verbose    bool
)

func init() { // Yeah yeah inits are bad go away
	flag.StringVar(&configPath, "config", os.ExpandEnv("$HOME/.config/tau.toml"), "sets the config file to use")
	flag.BoolVar(&verbose, "v", false, "enables verbose logging")

	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fmt.Printf("%s [FLAGS]... [FILE] [COMMAND]\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func fatalf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	os.Exit(1)
}

func main() {
	flag.Parse()

	args := flag.Args()

	if len(args) < 2 {
		flag.Usage()
		return
	}

	targetFile := args[0]
	commandToExecute := args[1]

	conf, err := getConf(configPath)
	if err != nil {
		fatalf("could not get config file: %s", err)
	}

	if targetFile == "" {
		if len(args) < 1 {
			fatalf("cannot transform nothing (file was unset)")
		}

		targetFile = args[0]
	}

	newName, err := doTransform(targetFile, conf)
	if err != nil {
		fatalf("could not transform path: %s", err)
	}

	if err := Execute(commandToExecute, targetFile, newName); err != nil {
		fatalf("error occurred while executing command: %s", err)
	}

}

func verboseLog(msgs ...interface{}) {
	if !verbose {
		return
	}
	fmt.Println(msgs...)
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
