package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/dotnews/conduit/pipeline"
	"github.com/dotnews/conduit/queue"
	"gopkg.in/yaml.v2"

	"github.com/golang/glog"
)

func main() {
	flag.Parse()
	glog.Info("Running conduit")

	for _, absPath := range findPipelines(os.Getenv("PIPELINE_ROOT")) {
		glog.Infof("Loading pipeline: %s", absPath)

		pipeline.New(
			filepath.Dir(absPath),
			loadPipeline(absPath),
			queue.New(1*time.Second),
		).Run()
	}

	select {}
}

// Find pipeline YAML files in base directory
func findPipelines(root string) []string {
	pipelines := []string{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && info.Name() == "pipeline.yml" {
			pipelines = append(pipelines, path)
		}

		return err
	})

	if err != nil {
		glog.Fatalf("Failed walking directory: %s", root)
	}

	return pipelines
}

// Load pipeline YAML file
func loadPipeline(file string) *pipeline.Meta {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		glog.Fatalf("Failed loading pipeline file: %s, error: %v", file, err)
	}

	var meta pipeline.Meta
	if err = yaml.Unmarshal(bytes, &meta); err != nil {
		glog.Fatalf("Failed parsing pipeline file: %s, error: %v", file, err)
	}

	return &meta
}
