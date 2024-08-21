---
id: package
title: Package
---

`ghz` can be used programmatically as Go package within Go applications. See detailed [godoc](https://godoc.org/github.com/chalk-ai/ghz) documentation. Example usage:


```go
package main

import (
	"fmt"
	"os"

	"github.com/chalk-ai/ghz/printer"
	"github.com/chalk-ai/ghz/runner"
)

func main() {
	report, err := runner.Run(
		"helloworld.Greeter.SayHello",
		"localhost:50051",
		runner.WithProtoFile("greeter.proto", []string{}),
		runner.WithDataFromFile("data.json"),
		runner.WithInsecure(true),
	)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	printer := printer.ReportPrinter{
		Out:    os.Stdout,
		Report: report,
	}

	printer.Print("pretty")
}
```
