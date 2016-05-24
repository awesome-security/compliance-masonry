package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/opencontrol/compliance-masonry/packages/config/common"
	"github.com/opencontrol/compliance-masonry/packages/config/parser"
	"github.com/opencontrol/compliance-masonry/packages/docs"
	"github.com/opencontrol/compliance-masonry/packages/docx"
	"github.com/opencontrol/compliance-masonry/packages/gitbook"
	"github.com/opencontrol/compliance-masonry/packages/tools/constants"
	"github.com/opencontrol/compliance-masonry/packages/tools/fs"
	"github.com/opencontrol/compliance-masonry/packages/tools/mapset"
)

var certification, exportPath, markdownPath, opencontrolDir, templatePath string

// NewCLIApp creates a new instances of the CLI
func NewCLIApp() *cli.App {
	app := cli.NewApp()
	app.Name = "Compliance Masonry"
	app.Usage = "Open Control CLI Tool"
	app.Version = "1.0.0"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "Indicates whether to run the command with verbosity.",
		},
	}
	app.Before = func(c *cli.Context) error {
		// Resets the log to output to nothing
		log.SetOutput(ioutil.Discard)
		if c.Bool("verbose") {
			log.SetOutput(os.Stderr)
			log.Println("Running with verbosity")
		}
		return nil
	}
	app.Commands = []cli.Command{
		{
			Name:    "get",
			Aliases: []string{"g"},
			Usage:   "Install compliance dependencies",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "dest",
					Value: constants.DefaultDestination,
					Usage: "Location to download the repos.",
				},
				cli.StringFlag{
					Name:  "config",
					Value: constants.DefaultConfigYaml,
					Usage: "Location of system yaml",
				},
			},
			Action: func(c *cli.Context) {
				f := fs.OSUtil{}
				config := c.String("config")
				configBytes, err := f.OpenAndReadFile(config)
				if err != nil {
					app.Writer.Write([]byte(err.Error()))
					os.Exit(1)
				}
				wd, err := os.Getwd()
				if err != nil {
					app.Writer.Write([]byte(err.Error()))
					os.Exit(1)
				}
				destination := filepath.Join(wd, c.String("dest"))
				err = Get(destination,
					configBytes,
					&common.ConfigWorker{Downloader: common.NewVCSDownloader(), Parser: parser.Parser{}, ResourceMap: mapset.Init(), FSUtil: f})
				if err != nil {
					app.Writer.Write([]byte(err.Error()))
					os.Exit(1)
				}
				app.Writer.Write([]byte("Compliance Dependencies Installed"))
			},
		},
		{
			Name:    "docs",
			Aliases: []string{"d"},
			Usage:   "Create Documentation",
			Subcommands: []cli.Command{
				{
					Name:    "gitbook",
					Aliases: []string{"g"},
					Usage:   "Create Gitbook Documentation",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "opencontrols, o",
							Value:       "opencontrols",
							Usage:       "Set opencontrols directory",
							Destination: &opencontrolDir,
						},
						cli.StringFlag{
							Name:        "exports, e",
							Value:       "exports",
							Usage:       "Sets the export directory",
							Destination: &exportPath,
						},
						cli.StringFlag{
							Name:        "markdowns, m",
							Value:       "markdowns",
							Usage:       "Sets the markdowns directory",
							Destination: &markdownPath,
						},
					},
					Action: func(c *cli.Context) {
						config := gitbook.Config{
							Certification:  c.Args().First(),
							OpencontrolDir: opencontrolDir,
							ExportPath:     exportPath,
							MarkdownPath:   markdownPath,
						}
						messages := docs.MakeGitbook(config)
						app.Writer.Write([]byte(strings.Join(messages, "\n")))
					},
				},
				{
					Name:    "docx",
					Aliases: []string{"d"},
					Usage:   "Create Docx Documentation using a Template",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "opencontrols, o",
							Value:       "opencontrols",
							Usage:       "Set opencontrols directory",
							Destination: &opencontrolDir,
						},
						cli.StringFlag{
							Name:        "template, t",
							Value:       "",
							Usage:       "Set template to build",
							Destination: &templatePath,
						},
						cli.StringFlag{
							Name:        "export, e",
							Value:       "export.docx",
							Usage:       "Sets the export directory",
							Destination: &exportPath,
						},
					},
					Action: func(c *cli.Context) {
						config := docx.Config{
							OpencontrolDir: opencontrolDir,
							TemplatePath:   templatePath,
							ExportPath:     exportPath,
						}
						messages := docs.BuildTemplate(config)
						app.Writer.Write([]byte(strings.Join(messages, "\n")))
					},
				},
			},
		},
		diffCommand,
	}
	return app
}

func main() {
	app := NewCLIApp()
	app.Run(os.Args)
}
