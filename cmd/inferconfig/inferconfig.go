package main

import (
	"fmt"
	"github.com/CircleCI-Public/circleci-config/generation"
	"github.com/CircleCI-Public/circleci-config/labeling"
	"github.com/CircleCI-Public/circleci-config/labeling/codebase"
	"log"
	"os"
)

var stderr = log.New(os.Stderr, "", 0)

func main() {
	dir := "."
	if len(os.Args) == 2 {
		dir = os.Args[1]
	}

	stat, err := os.Stat(dir)
	if len(os.Args) > 2 {
		stderr.Printf("usage: %s {path}", os.Args[0])
		os.Exit(1)
	}
	if os.IsNotExist(err) || !stat.IsDir() {
		stderr.Printf("%s is not a directory", dir)
		os.Exit(2)
	}
	if err != nil {
		stderr.Printf("error reading from %s: %v", os.Args, err)
		os.Exit(3)
	}

	cb := codebase.LocalCodebase{BasePath: dir}
	labels := labeling.ApplyAllRules(cb)
	config := generation.GenerateConfig(labels)
	fmt.Print(config)
}
