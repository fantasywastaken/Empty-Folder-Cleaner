# Empty-Folder-Cleaner

Recursively find and delete empty folders in any directory tree, including nested chains that become empty once their children are removed.

### How It Works

- Walks the given root once with `filepath.WalkDir`, building an in-memory tree of directories and file counts.
- Ignores hidden files and folders when `--ignore-hidden` is set, so `.DS_Store`, `.git`, and friends do not keep an otherwise empty tree alive.
- Traverses the tree bottom-up so nested empty chains (`a/b/c` where `c` is empty and so are `b` and `a`) are all collected in one pass.
- Sorts the collected paths deepest-first so removal never fails because a child still exists.
- Deletes only when `--delete` is set; otherwise prints the list so you can review or pipe it elsewhere.

## Setup

**Requirements**

- Go 1.21 or newer
- No external dependencies (standard library only)

**Installation**

```bash
git clone https://github.com/fantasywastaken/Empty-Folder-Cleaner.git
cd Empty-Folder-Cleaner
go build -o emptycleaner .
```

### Usage

Preview the empty folders under a project (no changes made):

```bash
$ emptycleaner ./project --dry-run
./project/build/tmp
./project/logs/2023
./project/logs

found 3 empty directories (use --delete to remove them)
```

Actually delete them:

```bash
$ emptycleaner ./project --delete
removed ./project/build/tmp
removed ./project/logs/2023
removed ./project/logs

removed 3 of 3 directories
```

Ignore hidden files so `.DS_Store` or `.gitkeep` do not save a folder:

```bash
$ emptycleaner ./project --delete --ignore-hidden
removed ./project/assets/.cache
removed ./project/assets
```

Verbose scanning of a big tree:

```bash
$ emptycleaner /var/log --dry-run --verbose
scanned 200 directories...
scanned 400 directories...
scanned 573 directories, found 12 empty
```

### Features

- Bottom-up traversal that catches nested empty chains in one pass.
- Optional hidden-file filtering with `--ignore-hidden`.
- Safe `--dry-run` preview mode before any destructive action.
- Progress output every 200 directories under `--verbose`.
- Mutually exclusive `--delete` and `--dry-run` flags to prevent accidents.
- Continues past unreadable directories with a warning instead of aborting.
- Pure standard library, no external dependencies.
