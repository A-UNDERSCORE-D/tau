package main

import (
	"fmt"
	"regexp"

	"github.com/pelletier/go-toml"
)

type transformConfig struct {
	Matcher         string
	compiledMatcher *regexp.Regexp `toml:"-"`
	Transform       string
}

type config map[string]*transformConfig

func (t *transformConfig) Compile() error {
	res, err := regexp.Compile(t.Matcher)
	if err != nil {
		return err
	}
	t.compiledMatcher = res
	return nil
}

func getConf(path string) (config, error) {
	tree, err := toml.LoadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could load config: %w", err)
	}

	out := make(config)
	if err := tree.Unmarshal(&out); err != nil {
		return nil, fmt.Errorf("could not unmarshal config: %w", err)
	}

	for name, conf := range out {
		if err := conf.Compile(); err != nil {
			return nil, fmt.Errorf("could not compile %q: %w", name, err)
		}

	}

	return out, nil
}
