package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/MrEthical07/superapi/internal/tools/routedocgen"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "routedocgen failed: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("routedocgen", flag.ContinueOnError)
	modulesRoot := fs.String("modules", "internal/modules", "path to module root")
	trackerPath := fs.String("tracker", "docs/ProjectBookDocs/endpoint-tracker.json", "path to endpoint tracker JSON")
	guidelinesPath := fs.String("guidelines", "docs/ProjectBookDocs/API-GUIDELINES.md", "path to API guidelines markdown")
	outputPath := fs.String("out", "docs/routeDetails.md", "output markdown file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	return routedocgen.Generate(routedocgen.Options{
		ModulesRoot:    *modulesRoot,
		TrackerPath:    *trackerPath,
		GuidelinesPath: *guidelinesPath,
		OutputPath:     *outputPath,
	})
}
