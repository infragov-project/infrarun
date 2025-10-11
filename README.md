# infrarun

A Go library and command-line tool for **running IaC analysis tools**.

---

## 🚀 Installation

You can install the CLI directly with:

```bash
go install github.com/infragov-project/infrarun@latest
```

Make sure your `$GOPATH/bin` (or Go’s default install path) is in your `PATH`.

To use the library in your Go project:

```bash
go get github.com/infragov-project/infrarun
```

Full API documentation is available on [pkg.go.dev](https://pkg.go.dev/github.com/infragov-project/infrarun).

---

## ⚙️ Usage

### Basic Example

```bash
infrarun run KICS
```

This runs [KICS](https://kics.io/) on the current working directory.

### Common Commands

| Command               | Description          |
| --------------------- | -------------------- |
| `infrarun run [name]` | Runs a given tool in the current working directory |
| `infrarun list`       | Lists all available tools|

You can also see available commands with:

```bash
infrarun help
```

or get command-specific help:

```bash
infrarun <command> --help
```

---

## 📚 Library Reference

If you’re using the Go library directly, see the full reference on
👉 [pkg.go.dev/github.com/infragov-project/infrarun](https://pkg.go.dev/github.com/infragov-project/infrarun)

Example usage:

```go
package main

import (
	"fmt"
	"os"

	"github.com/infragov-project/infrarun/pkg/infrarun"
)

func main() {

	tools := infrarun.GetAvailableTools()

	chosenTools := make([]*infrarun.Tool, 0)

	for n, t := range tools {
		fmt.Println(n, "-", t.Image())

		chosenTools = append(chosenTools, &t)
	}

	rep, err := infrarun.RunTools(chosenTools, ".")

	if err != nil {
		panic(err)
	}

	rep.PrettyWrite(os.Stdout)

}
```

---

## 🧪 Development

To build and test locally:

```bash
git clone https://github.com/infragov-project/infrarun
cd infrarun
go build
go test ./...
```

---

## 📄 License

This project is licensed under the [Apache License, Version 2.0](LICENSE).

---

## ✨ Contributing

Contributions are welcome!
Please open an issue or pull request on [GitHub](https://github.com/infragov-project/infrarun).
