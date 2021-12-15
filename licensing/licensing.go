package licensing

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

//go:embed my-licenses/*
var embeddedLicenses embed.FS

func PrintEmbeddedLicenses() {
	fmt.Print("Usage, modification, and distribution of this software and its components are subject to the following respective licensing terms:\n\n")
	_ = fs.WalkDir(embeddedLicenses, "my-licenses", MyWalkFunc)
}

func MyWalkFunc(path string, d fs.DirEntry, _ error) error {
	if !d.IsDir() && d.Name() != "make_embedFS_happy" { // Ignore directory entries, and the "make_embedFS_happy" file

		// This bit of code takes something like "my-licenses/github.com/jimmale/examplelicensesissue/LICENSE" and
		// yields "github.com/jimmale/examplelicensesissue"
		components := strings.Split(path, "/")
		components = components[1 : len(components)-1]
		cleanPath := filepath.Join(components...)

		// Read the content of the license file into a []byte
		fileContent, _ := fs.ReadFile(embeddedLicenses, path)

		fmt.Printf("License for %s:\n", cleanPath)
		fmt.Println("================================================================================")
		fmt.Print(string(fileContent))
		fmt.Println("================================================================================")
		fmt.Print("\n\n\n")
	}
	return nil
}
