package client

import (
	"fmt"

	"github.com/l2cup/kids1"
	"github.com/l2cup/kids1/pkg/color"
	"github.com/l2cup/kids1/pkg/dispatcher"
	"github.com/urfave/cli/v2"
)

func NewAddDir(app *kids1.App) *cli.Command {
	return &cli.Command{
		Name:  "ad",
		Usage: "Adds the directory to the crawler",
		Action: func(c *cli.Context) error {
			cErr := app.DirectoryCrawler.AddDirectoryPath(c.Args().Get(0))
			if cErr.IsNotNil() {
				fmt.Println(color.Red(cErr.Message))
				return nil
			}
			return nil
		},
	}
}

func NewGetFileSummary(app *kids1.App) *cli.Command {
	return &cli.Command{
		Name:  "file",
		Usage: "Gets file summaries",
		Action: func(c *cli.Context) error {
			results, err := app.ResultRetriever.GetSummaries(dispatcher.FileJobType)
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

func NewQueryFileCorpus(app *kids1.App) *cli.Command {
	return &cli.Command{
		Name:  "file",
		Usage: "Queries file corpuses",
		Action: func(c *cli.Context) error {
			results, err := app.ResultRetriever.QuerySummary(dispatcher.FileJobType, c.Args().Get(0))
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

func NewGetFileCorpus(app *kids1.App) *cli.Command {
	return &cli.Command{
		Name:  "file",
		Usage: "Gets file corpuses",
		Action: func(c *cli.Context) error {
			results, err := app.ResultRetriever.GetSummary(dispatcher.FileJobType, c.Args().Get(0))
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

func NewCFS(app *kids1.App) *cli.Command {
	return &cli.Command{
		Name:  "cfs",
		Usage: "clears file summary",
		Action: func(c *cli.Context) error {
			app.ResultRetriever.DeleteSummary(dispatcher.FileJobType)
			return nil
		},
	}
}
