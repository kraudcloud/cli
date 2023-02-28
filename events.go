package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/kraudcloud/cli/api"
	"github.com/mitchellh/colorstring"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"time"
)

func eventsCMD() *cobra.Command {

	c := &cobra.Command{
		Use:   "events",
		Short: "listen to cluster events",
		Run: func(cmd *cobra.Command, _ []string) {

			req, err := http.NewRequestWithContext(
				cmd.Context(),
				"GET",
				"/apis/kraudcloud.com/v1/events/stream.json",
				nil,
			)

			if err != nil {
				panic(err)
			}

			resp, err := API().DoRaw(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			scanner := bufio.NewScanner(resp.Body)

			for scanner.Scan() {

				if scanner.Text() == "" {
					continue
				}

				var ev api.KraudEvent

				err := json.Unmarshal(scanner.Bytes(), &ev)
				if err != nil {
					panic(err)
				}

				ts := ""
				if ev.Timestamp != nil {
					tst, _ := time.Parse(time.RFC3339, *ev.Timestamp)
					ts = tst.Format("2006-01-02 15:04:05")
				}

				resAndAction := ""
				if ev.Resource != nil {
					resAndAction = *ev.Resource
				}
				if ev.Action != nil {
					resAndAction = resAndAction + " " + *ev.Action
				}

				reason := ""
				if ev.Reason != nil {
					reason = *ev.Reason
				}

				var s = fmt.Sprintf("[cyan][%s] [blue]%-16s[reset] ",
					ts,
					resAndAction)

				if ev.Severity != nil && *ev.Severity == "error" {
					s += fmt.Sprintf("[red]%s[reset]\n", reason)
				} else if ev.Severity != nil && *ev.Severity == "warning" {
					s += fmt.Sprintf("[orange]%s[reset]\n", reason)
				} else {
					s += fmt.Sprintf("%s\n", reason)
				}

				if ev.System != nil {
					s += fmt.Sprintf("  system: %s\n", *ev.System)
				}

				if ev.ID != nil {
					s += fmt.Sprintf("  uuid: %s\n", *ev.ID)
				}

				if ev.AID != nil {
					s += fmt.Sprintf("  docker id: %s\n", *ev.AID)
				}

				if ev.TraceID != nil {
					s += fmt.Sprintf("  trace id: %s\n", *ev.TraceID)
				}

				if ev.UserID != nil {
					s += fmt.Sprintf("  user uuid: %s\n", *ev.UserID)
				}

				if ev.UserNR != nil {
					s += fmt.Sprintf("  user unix id: %d\n", *ev.UserNR)
				}

				if ev.Details != nil {
					s += fmt.Sprintf("  details: \n")
					for k, v := range *ev.Details {
						s += fmt.Sprintf("    %s: %v \n", k, v)
					}
				}

				s += "\n"

				colorstring.Fprintf(os.Stdout, s)
			}

		},
	}
	return c
}
