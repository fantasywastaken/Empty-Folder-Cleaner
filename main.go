package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type dirNode struct {
	path     string
	children []*dirNode
	files    int
}

func main() {
	dryRun := flag.Bool("dry-run", false, "List empty folders without deleting them")
	del := flag.Bool("delete", false, "Delete every empty folder that is found")
	verbose := flag.Bool("verbose", false, "Print scanning progress to stderr")
	ignoreHidden := flag.Bool("ignore-hidden", false, "Treat hidden files and folders as if they do not exist")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: emptycleaner [flags] <path>")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Recursively find and optionally delete empty folders.")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Flags:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	if *del && *dryRun {
		fmt.Fprintln(os.Stderr, "emptycleaner: --delete and --dry-run are mutually exclusive")
		os.Exit(2)
	}

	root := flag.Arg(0)
	info, err := os.Stat(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "emptycleaner:", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintln(os.Stderr, "emptycleaner: not a directory:", root)
		os.Exit(1)
	}

	empties, scanned, err := collectEmptyDirs(root, *ignoreHidden, *verbose)
	if err != nil {
		fmt.Fprintln(os.Stderr, "emptycleaner:", err)
		os.Exit(1)
	}

	sort.Slice(empties, func(i, j int) bool {
		return pathDepth(empties[i]) > pathDepth(empties[j])
	})

	if *verbose {
		fmt.Fprintf(os.Stderr, "scanned %d directories, found %d empty\n", scanned, len(empties))
	}

	if len(empties) == 0 {
		fmt.Println("no empty directories found")
		return
	}

	if *del {
		removed := 0
		for _, dir := range empties {
			if err := os.Remove(dir); err != nil {
				fmt.Fprintf(os.Stderr, "failed to remove %s: %v\n", dir, err)
				continue
			}
			removed++
			fmt.Printf("removed %s\n", dir)
		}
		fmt.Printf("\nremoved %d of %d directories\n", removed, len(empties))
		return
	}

	for _, dir := range empties {
		fmt.Println(dir)
	}
	fmt.Fprintf(os.Stderr, "\nfound %d empty directories (use --delete to remove them)\n", len(empties))
}

func pathDepth(p string) int {
	return strings.Count(p, string(os.PathSeparator))
}

func collectEmptyDirs(root string, ignoreHidden, verbose bool) ([]string, int, error) {
	nodes := make(map[string]*dirNode)
	rootNode := &dirNode{path: root}
	nodes[root] = rootNode

	scanned := 0
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", walkErr)
			return nil
		}

		if d.IsDir() {
			scanned++
			if verbose && scanned%200 == 0 {
				fmt.Fprintf(os.Stderr, "scanned %d directories...\n", scanned)
			}
			if path == root {
				return nil
			}
			base := filepath.Base(path)
			if ignoreHidden && strings.HasPrefix(base, ".") {
				return filepath.SkipDir
			}
			node := &dirNode{path: path}
			nodes[path] = node
			if parent := nodes[filepath.Dir(path)]; parent != nil {
				parent.children = append(parent.children, node)
			}
			return nil
		}

		if ignoreHidden && strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		if parent := nodes[filepath.Dir(path)]; parent != nil {
			parent.files++
		}
		return nil
	})
	if err != nil {
		return nil, scanned, err
	}

	var empties []string
	var visit func(n *dirNode) bool
	visit = func(n *dirNode) bool {
		empty := n.files == 0
		for _, ch := range n.children {
			if !visit(ch) {
				empty = false
			}
		}
		if empty && n != rootNode {
			empties = append(empties, n.path)
		}
		return empty
	}
	visit(rootNode)

	return empties, scanned, nil
}
