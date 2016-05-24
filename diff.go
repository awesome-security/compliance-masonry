package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/opencontrol/compliance-masonry/packages/inventory"
	"github.com/tg/gosortmap"
	"strings"
)

const (
	diffCommandName  = "diff"
	diffCommandUsage = "Compute Gap Analysis"
)

var (
	diffCommandAliases = []string{"d"}
	diffCommandFlags   = []cli.Flag{
		cli.StringFlag{
			Name:        "opencontrols, o",
			Value:       "opencontrols",
			Usage:       "Set opencontrols directory",
			Destination: &opencontrolDir,
		},
	}
	diffCommand = cli.Command{
		Name:    diffCommandName,
		Aliases: diffCommandAliases,
		Usage:   diffCommandUsage,
		Flags:   diffCommandFlags,
		Action:  diffCommandAction,
	}
)

func diffCommandAction(c *cli.Context) error {
	config := inventory.Config{
		Certification:  c.Args().First(),
		OpencontrolDir: opencontrolDir,
	}
	inventory, err := inventory.ComputeGapAnalysis(config)
	if err != nil && len(err) > 0 {
		return cli.NewExitError(strings.Join(err, "\n"), 1)
	}

	c.App.Writer.Write([]byte(fmt.Sprintf("\nNumber of missing controls: %d\n", len(inventory.MissingControlList))))
	for _, standardAndControl := range sortmap.ByKey(inventory.MissingControlList) {
		c.App.Writer.Write([]byte(fmt.Sprintf("%s\n", standardAndControl.Key)))
	}
	return nil
}
