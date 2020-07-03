/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/m00nf1sh/golang-license/pkg/licensee"
	"github.com/m00nf1sh/golang-license/pkg/module"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strings"
)

func newAnalysisCmd(rootCmd *cobra.Command) *analysisCmd {
	ac := &analysisCmd{
		cmd: &cobra.Command{
			Use:   "analysis",
			Short: "analysis license of golang packages",
			Long:  `analysis license of golang packages and it's direct or indirect dependencies'`,
		},
	}
	ac.cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return ac.run(context.Background())
	}
	rootCmd.AddCommand(ac.cmd)
	ac.cmd.Flags().StringVar(&ac.pattern, "pattern", "./...", "pattern for your package")
	return ac
}

// analysisCmd represents the analysis command
type analysisCmd struct {
	cmd     *cobra.Command
	pattern string
}

type DependencyInfo struct {
	Module             string `json:"module"`
	Version            string `json:"version"`
	Repository         string `json:"repository"`
	License            string `json:"license"`
	LicenseAttribution string `json:"licenseAttribution"`
	LicenseContent     string `json:"licenseContent"`
}

type AnalysisOutput struct {
	Dependencies []DependencyInfo `json:"dependencies"`
}

func (ac *analysisCmd) run(ctx context.Context) error {
	dependencyResolver := module.NewDependencyResolver()
	repositoryLocator := module.NewRepositoryLocator()
	licenseDetector := licensee.NewDetector()

	analysisOutput := AnalysisOutput{}
	dependencies, err := dependencyResolver.Resolve(ctx, ac.pattern)
	if err != nil {
		return err
	}

	for _, dependency := range dependencies {
		if dependency.Module.Main {
			continue
		}

		detectRet, err := licenseDetector.Detect(dependency)
		if err != nil {
			return err
		}
		repository, err := repositoryLocator.Locate(dependency.Module)
		if err != nil {
			glog.Error(errors.Wrapf(err, "failed to locate repository for %v", dependency.Module.Path))
		}

		var license string
		var licenseAttribution string
		var licenseContent string
		licenseFile, err := ac.filterLicenseFiles(detectRet)
		if err != nil {
			license = "UNKNOWN"
			glog.Error(errors.Wrapf(err, "failed to filter licenses for %v", dependency.Module.Path))
		} else {
			license = licenseFile.MatchedLicense
			licenseAttribution = licenseFile.Attribution
			licenseContent = licenseFile.Content
		}
		dependencyInfo := DependencyInfo{
			Module:             dependency.Module.Path,
			Version:            dependency.Module.Version,
			Repository:         repository,
			License:            license,
			LicenseAttribution: licenseAttribution,
			LicenseContent:     licenseContent,
		}
		analysisOutput.Dependencies = append(analysisOutput.Dependencies, dependencyInfo)
	}
	payload, _ := json.Marshal(analysisOutput)
	fmt.Println(string(payload))
	return nil
}

func (ac *analysisCmd) filterLicenseFiles(detectRet licensee.DetectionResult) (licensee.MatchedFile, error) {
	var matchedFiles []licensee.MatchedFile
	var filenames []string
	for _, matchedFile := range detectRet.MatchedFiles {
		filenames = append(filenames, matchedFile.Filename)
		if matchedFile.MatchedLicense != "NOASSERTION" {
			matchedFiles = append(matchedFiles, matchedFile)
		}
	}

	if len(matchedFiles) == 0 {
		return licensee.MatchedFile{}, errors.Errorf("no matching license found, please check files: %v", strings.Join(filenames, ","))
	}
	if len(matchedFiles) > 1 {
		return licensee.MatchedFile{}, errors.Errorf("multiple matching license found, please check files: %v", strings.Join(filenames, ","))
	}
	return matchedFiles[0], nil
}
