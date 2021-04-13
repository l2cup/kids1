package main

import (
	"fmt"
	"strings"

	"io"

	"github.com/bobappleyard/readline"
	"github.com/l2cup/kids1"
	"github.com/l2cup/kids1/cmd/client"
	"github.com/l2cup/kids1/pkg/color"
	"github.com/urfave/cli/v2"
)

func main() {
	app := kids1.New()
	cmd := createCmd(app)

	app.Start()

	startLoop(cmd, app)
}

func createCmd(app *kids1.App) *cli.App {
	cmd := cli.NewApp()

	cmd.Name = "map-reduce wc"
	cmd.Description = "kids1 kids1 kids1 kids1"
	cmd.UsageText = "command [command options] [arguments...]"
	cmd.Authors = append(cmd.Authors, &cli.Author{Name: "l2cup", Email: "nikolic.uros@me.com"})
	cmd.CommandNotFound = func(c *cli.Context, command string) {
		fmt.Println(color.Red(fmt.Sprintf("no matching command '%s'", command)))
		cli.ShowAppHelp(c)
	}

	cmd.OnUsageError = func(context *cli.Context, err error, isSubcommand bool) error {
		fmt.Println("incorrect usage")
		return nil
	}

	cmd.Commands = []*cli.Command{
		client.NewAddDir(app),
		client.NewAddWeb(app),
		client.NewGet(app),
		client.NewQuery(app),
		client.NewSummary(app),
		client.NewCFS(app),
		client.NewCWS(app),
	}

	return cmd
}

func startLoop(cmd *cli.App, app *kids1.App) {
	for {
		line, err := readline.String(color.Yellow("$ "))
		if err == io.EOF {
			break

		}
		if err != nil {
			fmt.Println("error: ", err)
			break

		}
		readline.AddHistory(line)
		if strings.ToLower(line) == "exit" || strings.ToLower(line) == "stop" {
			app.Stop()
			break
		}

		err = cmd.Run(strings.Fields("cmd " + line))

		if err != nil {
			fmt.Println(err)
		}
	}
}
