package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
)

var args Arguments

type Arguments struct {
	DryRun bool `short:"d" long:"dryrun" description:"List changes that would made instead of making them"`
	Unstub bool `short:"r" long:"remove" description:"Unstub a folder and it's subfolders (deletes all x_test.go)"`
	Pos    struct {
		BaseDir string `positional-arg-name:"DIR"`
	} `positional-args:"yes"`
}

func main() {
	_, err := flags.Parse(&args)
	if flags.WroteHelp(err) {
		os.Exit(0)
	} else if err != nil {
		panic(err)
	}

	baseDir := `.`
	if len(args.Pos.BaseDir) != 0 {
		baseDir = args.Pos.BaseDir
	}

	folders := getSubDirs(baseDir)
	for i := 0; i < len(folders); i++ {
		subs := getSubDirs(folders[i])
		folders = append(folders, subs...)
	}

	for i := 0; i < len(folders); i++ {
		if !args.Unstub && !validFolder(folders[i]) {
			folders = append(folders[:i], folders[i+1:]...)
			i--
		} else if args.Unstub && !hasXTest(folders[i]) {
			folders = append(folders[:i], folders[i+1:]...)
			i--
		}
	}

	if !args.Unstub && validFolder(baseDir) {
		filename := filepath.Join(baseDir, "x_test.go")
		if args.DryRun {
			fmt.Printf("create %s\n", filename)
		} else {
			err := ioutil.WriteFile(filename, []byte("package main\n"), 0666)
			if err != nil {
				fmt.Printf("unable to write %s: %+v", filename, err)
			}
		}
	} else if args.Unstub {
		filename := filepath.Join(baseDir, "x_test.go")
		if args.DryRun {
			fmt.Printf("delete %s\n", filename)
		} else {
			err := os.Remove(filename)
			if err != nil {
				fmt.Printf("unable to delete %s: %+v", filename, err)
			}
		}
	}

	for _, f := range folders {
		filename := filepath.Join(f, "x_test.go")
		if args.Unstub {
			if args.DryRun {
				fmt.Printf("delete %s\n", filename)
			} else {
				err := os.Remove(filename)
				if err != nil {
					fmt.Printf("unable to delete %s: %+v", filename, err)
				}
			}
		} else {
			if args.DryRun {
				fmt.Printf("create %s\n", filename)
			} else {
				pkg := "package " + filepath.Base(f)
				err := ioutil.WriteFile(filename, []byte(pkg), 0666)
				if err != nil {
					fmt.Printf("unable to write %s: %+v", filename, err)
				}
			}
		}
	}
}

func validFolder(dir string) (valid bool) {
	containsGoFiles := false
	containsTestFiles := false

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		filename := strings.ToUpper(f.Name())
		if strings.HasSuffix(filename, ".GO") {
			containsGoFiles = true
		}

		if strings.HasSuffix(filename, "_TEST.GO") {
			containsTestFiles = true
		}
	}

	return containsGoFiles && !containsTestFiles
}

func hasXTest(dir string) (has bool) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		if !f.IsDir() && f.Name() == "x_test.go" {
			return true
		}
	}

	return false
}

func getSubDirs(dir string) (subs []string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		absPath := filepath.Join(dir, f.Name())
		if f.IsDir() && f.Name()[0] != '.' && !contains(subs, absPath) {
			subs = append(subs, absPath)
		}
	}

	return subs
}

func contains(arr []string, x string) (c bool) {
	for _, a := range arr {
		if a == x {
			return true
		}
	}

	return false
}
