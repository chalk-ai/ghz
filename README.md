Fork of the [excellent ghz package](https://github.com/chalk-ai/ghz/tree/master) with some small modifications. This version improves the histogram in the report and fixes some small bugs.

# ghz

<div align="center">
	<br>
	<img src="green_fwd2.svg" alt="Logo" width="100">
	<br>
</div>

[![Release](https://img.shields.io/github/release/bojand/ghz.svg?style=flat-square)](https://github.com/chalk-ai/ghz/releases/latest)
![Build Status](https://github.com/chalk-ai/ghz/workflows/build/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/chalk-ai/ghz?style=flat-square)](https://goreportcard.com/report/github.com/chalk-ai/ghz)
[![License](https://img.shields.io/github/license/bojand/ghz.svg?style=flat-square)](https://raw.githubusercontent.com/bojand/ghz/master/LICENSE)
[![Donate](https://img.shields.io/badge/Donate-PayPal-green.svg?style=flat-square)](https://www.paypal.me/bojandj)
[![Buy me a coffee](https://img.shields.io/badge/buy%20me-a%20coffee-orange.svg?style=flat-square)](https://www.buymeacoffee.com/bojand)

[gRPC](http://grpc.io/) benchmarking and load testing tool.

## Documentation

All documentation for the library fork on [pkg.go.dev](https://pkg.go.dev/github.com/chalk-ai/ghz).


## Go Package

```go
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
```

## Credit

Icon made by <a href="http://www.freepik.com" title="Freepik">Freepik</a> from <a href="https://www.flaticon.com/" title="Flaticon">www.flaticon.com</a> is licensed by <a href="http://creativecommons.org/licenses/by/3.0/" title="Creative Commons BY 3.0" target="_blank">CC 3.0 BY</a>

## License

Apache-2.0
