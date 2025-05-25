# 📁 Files Merger

> *A simple yet powerful CLI utility to merge files recursively from directories into a single output file with their names and contents.*

## 🧾 Description

`files_merger` is a command-line utility that recursively walks through specified directories, finds files with given extensions, adds a comment line with the file path before its content, and merges everything into a single output file.

It's ideal for:
- Collecting source code from multiple files.
- Generating documentation or reports with context.
- Quickly reviewing the contents of a project directory.

---

## ⚙️ Installation

### Install via `go install`
```bash
go install github.com/trofkm/files_merger@latest
```

---

## 🛠️ Usage

```bash
./files_merger [flags] path...
```

### Example:
```bash
./files_merger -extensions=go,mod -ignore=.git -comment="//" ./src ./pkg
```

---

## 📌 Available Flags

| Flag           | Default             | Description |
|----------------|---------------------|-------------|
| `-extensions`  | `go`                | Comma-separated list of file extensions to process (e.g., `go,js,txt`). |
| `-output`      | `output.txt`        | Output file name. |
| `-ignore`      | `.git,.idea`        | Comma-separated list of directory names to ignore (e.g., `.git,node_modules`). |
| `-comment`     | `//`                | Comment symbol used to prefix file paths in the output. |

---

## 🔍 Sample Output

The output file (`output.txt`) will look like this:

```txt
// main.go
package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}

// go.mod
module example.com/myproject

go 1.21
```

---

## 🧪 Example Commands

### Merge all `.go` files in current directory:
```bash
./files_merger .
```

### Merge `.go` and `.mod` files, ignoring `.git`:
```bash
./files_merger -extensions=go,mod -ignore=.git .
```

### Save output to a custom file:
```bash
./files_merger -output=merged_code.txt .
```

---

## 💡 How It Works

1. The program recursively traverses all provided directories.
2. Skips directories listed in the `-ignore` flag using exact match.
3. For each matching file (based on extension), it writes:
    - A comment line with the file path.
    - The full content of the file.
4. All data is collected and written to the output file concurrently using goroutines and channels.
