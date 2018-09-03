package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/jspawar/generate-bazel-workspace-gradle/maven"
)

func main() {
	// run `gradle install`

	// TODO: change to just grab path to pom from path to `build.gradle`
	// parse generated `pom.xml`
	pomPath := os.Args[1]
	pomStr, err := ioutil.ReadFile(pomPath)
	if err != nil {
		panic(err)
	}

	pom := &maven.MavenPom{}
	err = xml.Unmarshal(pomStr, pom)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", pom)
}
