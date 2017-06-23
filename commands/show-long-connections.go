package commands

import (
	"fmt"
	"os"
	"strconv"
	"text/template"

	"github.com/ocmdev/rita/database"
	"github.com/ocmdev/rita/datatypes/data"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

func init() {
	command := cli.Command{

		Name:  "show-long-connections",
		Usage: "Print long connections and relevant information",
		Flags: []cli.Flag{
			humanFlag,
			databaseFlag,
			configFlag,
		},
		Action: func(c *cli.Context) error {
			if c.String("database") == "" {
				return cli.NewExitError("Specify a database with -d", -1)
			}

			res := database.InitResources(c.String("config"))

			var longConns []data.Conn
			coll := res.DB.Session.DB(c.String("database")).C(res.System.StructureConfig.ConnTable)

			sortStr := "-duration"

			coll.Find(nil).Sort(sortStr).All(&longConns)

			if len(longConns) == 0 {
				return cli.NewExitError("No results were found for "+c.String("database"), -1)
			}

			if c.Bool("human-readable") {
				err := showConnsHuman(longConns)
				if err != nil {
					return cli.NewExitError(err.Error(), -1)
				}
			}
			err := showConns(longConns)
			if err != nil {
				return cli.NewExitError(err.Error(), -1)
			}
			return nil
		},
	}
	bootstrapCommands(command)
}

func showConns(connResults []data.Conn) error {
	tmpl := "{{.Src}},{{.Spt}},{{.Dst}},{{.Dpt}},{{.Dur}},{{.Proto}}\n"

	out, err := template.New("Conn").Parse(tmpl)
	if err != nil {
		return err
	}

	for _, result := range connResults {
		err := out.Execute(os.Stdout, result)
		if err != nil {
			fmt.Fprintf(os.Stdout, "ERROR: Template failure: %s\n", err.Error())
		}
	}
	return nil
}

func showConnsHuman(connResults []data.Conn) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Source IP", "Source Port", "Destination IP",
		"Destination Port", "Duration", "Protocol"})
	for _, result := range connResults {
		table.Append([]string{
			result.Src,
			strconv.Itoa(result.Spt),
			result.Dst,
			strconv.Itoa(result.Dpt),
			strconv.FormatFloat(result.Dur, 'f', 2, 64),
			result.Proto,
		})
	}
	table.Render()
	return nil
}
