package client

import (
	"fmt"

	"github.com/l2cup/kids1"
	"github.com/l2cup/kids1/pkg/color"
	"github.com/l2cup/kids1/pkg/dispatcher"
	"github.com/urfave/cli/v2"
)

func NewAddWeb(app *kids1.App) *cli.Command {
	return &cli.Command{
		Name:  "aw",
		Usage: "Adds the url to the crawler",
		Action: func(c *cli.Context) error {
			app.WebCrawler.AddWebPage(c.Args().Get(0))
			return nil
		},
	}
}

func NewGetWebSummary(app *kids1.App) *cli.Command {
	return &cli.Command{
		Name:  "web",
		Usage: "Gets web summaries",
		Action: func(c *cli.Context) error {
			results, err := app.ResultRetriever.GetSummaries(dispatcher.WebJobType)
			if err != nil {
				fmt.Println(color.Red(err))
				return nil
			}
			if results == nil {
				fmt.Println(color.Red("results for summary do not exist"))
				return nil
			}

			for corpusName, summary := range results {
				fmt.Printf("[%s]\n", fmt.Sprint(color.Info(corpusName)))
				for k, v := range summary {
					fmt.Printf("%s: %d\n", fmt.Sprint(color.Purple(k)), v)
				}
			}
			return nil
		},
	}
}

func NewQueryWebCorpus(app *kids1.App) *cli.Command {
	return &cli.Command{
		Name:  "web",
		Usage: "Queries web corpuses",
		Action: func(c *cli.Context) error {
			results, err := app.ResultRetriever.QueryWebSummary(c.Args().Get(0))
			if err != nil {
				fmt.Println(color.Red(err))
				return nil
			}
			if results == nil {
				fmt.Println(color.Red("results for summary are not ready yet"))
				return nil
			}
			fmt.Println(color.Yellow("Printing results for corpus: %s\n", c.Args().Get(0)))
			for k, v := range results {
				fmt.Printf("%s : %d\n", fmt.Sprint(color.Info(k)), v)
			}
			return nil
		},
	}
}

func NewGetWebCorpus(app *kids1.App) *cli.Command {
	return &cli.Command{
		Name:  "web",
		Usage: "Gets web corpuses",
		Action: func(c *cli.Context) error {
			results, err := app.ResultRetriever.GetSummary(dispatcher.WebJobType, c.Args().Get(0))
			if err != nil {
				fmt.Println(color.Red(err))
				return nil
			}
			if results == nil {
				fmt.Println(color.Red("results for summary do not exist"))
				return nil
			}
			fmt.Println(color.Yellow("Printing results for corpus: %s\n", c.Args().Get(0)))
			for k, v := range results {
				fmt.Printf("%s : %d\n", fmt.Sprint(color.Info(k)), v)
			}
			return nil
		},
	}
}

func NewCWS(app *kids1.App) *cli.Command {
	return &cli.Command{
		Name:  "cws",
		Usage: "clears web summary",
		Action: func(c *cli.Context) error {
			app.ResultRetriever.DeleteSummary(dispatcher.WebJobType)
			return nil
		},
	}
}
