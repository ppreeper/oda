/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ppreeper/odoojrpc"
	"github.com/spf13/cobra"
)

var q QueryDef

var queryCmd = &cobra.Command{
	Use:     "query",
	Short:   "Query an Odoo model",
	Long:    `Query an Odoo model`,
	GroupID: "database",
	Run: func(cmd *cobra.Command, args []string) {
		if !IsProject() {
			return
		}

		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "no model specified")
			return
		}
		q.Model = args[0]

		cwd, project := GetProject()
		instance, err := GetInstance(project)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error getting instance %w", err)
			return
		}

		dbname := GetOdooConf(cwd, "db_name")

		oc := odoojrpc.NewOdoo().
			WithHostname(instance.IP4).
			WithPort(8069).
			WithDatabase(dbname).
			WithUsername(q.Username).
			WithPassword(q.Password).
			WithSchema("http")

		err = oc.Login()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error logging in %w", err)
			return
		}

		umdl := strings.Replace(q.Model, "_", ".", -1)

		fields := parseFields(q.Fields)
		if q.Count {
			fields = []string{"id"}
		}

		filtp, err := parseFilter(q.Filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err.Error())
		}

		rr, err := oc.SearchRead(umdl, filtp, q.Offset, q.Limit, fields)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err.Error())
		}
		if q.Count {
			fmt.Fprintln(os.Stderr, "records:", len(rr))
		} else {
			jsonStr, err := json.MarshalIndent(rr, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err.Error())
			}
			fmt.Fprintln(os.Stderr, string(jsonStr))
		}
	},
}

func init() {
	rootCmd.AddCommand(queryCmd)
	queryCmd.Flags().StringVarP(&q.Filter, "filter", "d", "", "domain filter")
	queryCmd.Flags().IntVarP(&q.Offset, "offset", "o", 0, "offset")
	queryCmd.Flags().IntVarP(&q.Limit, "limit", "l", 0, "limit records returned")
	queryCmd.Flags().StringVarP(&q.Fields, "fields", "f", "", "fields to return")
	queryCmd.Flags().BoolVarP(&q.Count, "count", "c", false, "count records")
	queryCmd.Flags().StringVarP(&q.Username, "username", "u", "admin", "username")
	queryCmd.Flags().StringVarP(&q.Password, "password", "p", "admin", "password")
}
