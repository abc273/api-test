package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var routePattern = regexp.MustCompile(`\.(GET|POST|PUT|PATCH|DELETE)\("([^"]+)"`)

func main() {
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	routerDir := filepath.Join(root, "router")
	var endpoints []string
	err = filepath.WalkDir(routerDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			matches := routePattern.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				endpoints = append(endpoints, fmt.Sprintf("%s %s", match[1], match[2]))
			}
		}
		return scanner.Err()
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	sort.Strings(endpoints)
	for _, endpoint := range endpoints {
		fmt.Println(endpoint)
	}
}
