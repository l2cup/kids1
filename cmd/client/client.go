package client

import (
	"github.com/l2cup/kids1"
	"github.com/urfave/cli/v2"
)

func NewSummary(app *kids1.App) *cli.Command {
	return &cli.Command{
		Name:  "summary",
		Usage: "returns corpus type summaries",
		Subcommands: []*cli.Command{
			NewGetFileSummary(app),
			NewGetWebSummary(app),
		},
	}
}

func NewGet(app *kids1.App) *cli.Command {
	return &cli.Command{
		Name:  "get",
		Usage: "Gets corpus results",
		Subcommands: []*cli.Command{
			NewGetFileCorpus(app),
			NewGetWebCorpus(app),
		},
	}
}

func NewQuery(app *kids1.App) *cli.Command {
	return &cli.Command{
		Name:  "query",
		Usage: "Queries corpus results",
		Subcommands: []*cli.Command{
			NewQueryFileCorpus(app),
			NewQueryWebCorpus(app),
		},
	}
}
