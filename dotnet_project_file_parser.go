package dotnetpublish

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
)

type ProjectFileParser struct{}

func NewProjectFileParser() ProjectFileParser {
	return ProjectFileParser{}
}

func (p ProjectFileParser) ASPNetIsRequired(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer file.Close()

	var project struct {
		SDK        string `xml:"Sdk,attr"`
		ItemGroups []struct {
			PackageReferences []struct {
				Include string `xml:"Include,attr"`
				Version string `xml:"Version,attr"`
			} `xml:"PackageReference"`
		} `xml:"ItemGroup"`
	}

	err = xml.NewDecoder(file).Decode(&project)
	if err != nil {
		return false, fmt.Errorf("failed to decode %s: %w", path, err)
	}

	if project.SDK == "Microsoft.NET.Sdk.Web" {
		return true, nil
	}

	for _, ig := range project.ItemGroups {
		for _, pr := range ig.PackageReferences {
			if pr.Include == "Microsoft.AspNetCore.App" || pr.Include == "Microsoft.AspNetCore.All" {
				return true, nil
			}
		}
	}
	return false, nil
}

func (p ProjectFileParser) NodeIsRequired(path string) (bool, error) {
	needsNode, err := findInFile("node ", path)
	if err != nil {
		return false, err
	}

	needsNPM, err := findInFile("npm ", path)
	if err != nil {
		return false, err
	}

	return needsNode || needsNPM, nil
}

func (p ProjectFileParser) NPMIsRequired(path string) (bool, error) {
	return findInFile("npm ", path)
}

func findInFile(str, path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer file.Close()

	var project struct {
		Targets []struct {
			Execs []struct {
				Command string `xml:",attr"`
			} `xml:"Exec"`
		} `xml:"Target"`
	}

	err = xml.NewDecoder(file).Decode(&project)
	if err != nil {
		return false, fmt.Errorf("failed to decode %s: %w", path, err)
	}

	for _, target := range project.Targets {
		for _, exec := range target.Execs {
			if strings.HasPrefix(exec.Command, str) {
				return true, nil
			}
		}
	}

	return false, nil
}
