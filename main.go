package main

import (
	"bufio"
	"fmt"
	"io"
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
	// parse flags
	_, err := flags.Parse(&args)
	if flags.WroteHelp(err) {
		os.Exit(0)
	} else if err != nil {
		panic(err)
	}

	args.DryRun = true

	// figure out target directory, assume current directory if no argument is passed in
	baseDir := `.`
	if len(args.Pos.BaseDir) != 0 {
		baseDir = args.Pos.BaseDir
	}

	// get all sub directories
	folders := []string{baseDir}
	for i := 0; i < len(folders); i++ {
		subs := getSubDirs(folders[i])
		folders = append(folders, subs...)
	}

	// filter sub directories
	for i := 0; i < len(folders); i++ {
		if (args.Unstub && !hasXTest(folders[i])) || (!args.Unstub && !validFolder(folders[i])) {
			folders = append(folders[:i], folders[i+1:]...)
			i--
		}
	}

	// iterate folders stubbing or unstubbing
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
			pkg := getPackageFromFolder(f)
			if args.DryRun {
				fmt.Printf("create %s with package %s\n", filename, pkg)
			} else {
				pkg := "package " + pkg
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

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		filename := strings.ToUpper(f.Name())
		if strings.HasSuffix(filename, ".GO") {
			containsGoFiles = true
		}

		if strings.HasSuffix(filename, "_TEST.GO") {
			// found a test file, nothing else matters,
			// we don't want to touch this folder
			return false
		}
	}

	return containsGoFiles
}

func hasXTest(dir string) (has bool) {
	// determines if a folder has a x_test.go file already in it
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

func getPackageFromFolder(dir string) (pkg string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	pkgs := map[string]int{} // keeps track of how many times a pkg is seen
	for _, f := range files {
		F := strings.ToUpper(f.Name())
		if f.IsDir() || !strings.HasSuffix(F, ".GO") || strings.HasSuffix(F, "_TEST.GO") {
			continue
		}
		pkg := getPackageFromSource(filepath.Join(dir, f.Name()))
		pkgs[pkg]++
	}

	// determine which package was seen the most
	most := ""
	for k, v := range pkgs {
		if v > pkgs[most] {
			most = k
		}
	}

	return most
}

func getPackageFromSource(path string) (pkg string) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("unable to retrieve package from %s: %+v", path, err)

		return ""
	}

	r := bufio.NewReader(f)

	var line string

	for err != io.EOF {
		// read a line
		line, err = r.ReadString('\n')
		if err != nil {
			fmt.Printf("unable to read from %s: %+v\n", path, err)
			return ""
		}

		// remove any comments
		commentStart := strings.Index(line, "//")
		if commentStart >= 0 {
			line = line[:commentStart]
		}

		// check for package
		if strings.HasPrefix(line, "package ") {
			// found the package, done!
			return strings.TrimSpace(strings.TrimPrefix(line, "package "))
		}
	}

	return ""
}

func getSubDirs(dir string) (subs []string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		absPath := filepath.Join(dir, f.Name())
		// filter to only dirs, no dot dirs (i.e. .git), and no duplicates (might happen with symlinked folders)
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
