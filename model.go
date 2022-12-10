package main

import (
	"strings"

	"github.com/rivo/tview"
)

// value with an ID and a name
type Value struct {
	ID   string
	Name string
}

// value with an ID, a name and a description
type ValueWithDesc struct {
	ID          string
	Name        string
	Description string
}

// alias to sort an array of ValueWithDesc by name
type ValueWithDescByName []ValueWithDesc

func (v ValueWithDescByName) Len() int {
	return len(v)
}

func (v ValueWithDescByName) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v ValueWithDescByName) Less(i, j int) bool {
	return strings.Compare(v[i].Name, v[j].Name) == -1
}

// application state, including spring initializer's options
type AppState struct {
	App                    *tview.Application
	Pages                  *tview.Pages
	DefaultGroupId         string
	DefaultArtifactId      string
	DefaultVersion         string
	DefaultName            string
	DefaultDescription     string
	DefaultPackageName     string
	Dependency             map[string][]ValueWithDesc
	SpringBuildTools       []ValueWithDesc
	DefaultSpringBuildTool int
	Packaging              []Value
	DefaultPackaging       int
	Languages              []Value
	DefaultLanguage        int
	SpringVersions         []Value
	DefaultSpringVersion   int
	JavaVersions           []Value
	DefaultJavaVersion     int
}

// application data, contains the values used to generate the project package
type AppData struct {
	Group             string
	Artifact          string
	Name              string
	Description       string
	Pkg               string
	SpringBuildTool   string
	Language          string
	JavaVersion       string
	SpringBootVersion string
	Packaging         string
	Dependencies      []ValueWithDesc
	Path              string
}
