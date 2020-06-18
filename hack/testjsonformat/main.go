/*
Copyright 2020 The KubeCarrier Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type TestEvent struct {
	Time    time.Time // encodes as an RFC3339-format string
	Action  string
	Package string
	Test    string
	Elapsed float64 // seconds
	Output  string
}

func main() {
	if err := (&cobra.Command{
		Use:  "testjsonformat",
		Long: "testjsonformat takes jsonline produced by go tool test2json and rearranges parallel test in proper ordering after the fact",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			in := cmd.InOrStdin()
			if len(args) == 1 {
				f, err := os.Open(args[0])
				if err != nil {
					return err
				}
				in = f
				defer f.Close()
			}
			reader := bufio.NewReader(in)
			keys := make([]string, 0)
			output := make(map[string]*strings.Builder)
			pf := make([]string, 0)
			lastTest := ""

			for {
				line, err := reader.ReadBytes('\n')
				if len(line) > 0 {
					ev := &TestEvent{}
					if err := json.Unmarshal(line, ev); err != nil {
						panic(err)
					}

					if ev.Action == "output" {
						if strings.HasPrefix(strings.TrimSpace(ev.Output), "--- PASS:") ||
							strings.HasPrefix(strings.TrimSpace(ev.Output), "--- FAIL:") {
							pf = append(pf, ev.Output)
							continue
						}

						ev.Test = lastTest
						// go tool test2json is a buggy attributing the right line to the right test
						candidate := strings.SplitN(strings.TrimSpace(ev.Output), ":", 2)[0]
						if _, ok := output[candidate]; strings.HasPrefix(ev.Output, "    ") &&
							strings.Count(ev.Output, ":") >= 2 &&
							ok {
							ev.Test = candidate
						}
					}

					if strings.Contains(ev.Output, "=== RUN") {
						if _, ok := output[ev.Test]; ok {
							panic("only single " + ev.Output + " should be here")
						}
						output[ev.Test] = new(strings.Builder)
						keys = append(keys, ev.Test)
					}

					switch ev.Action {
					case "output":
						if _, err := output[ev.Test].WriteString(ev.Output); err != nil {
							return err
						}
					}
					lastTest = ev.Test
				}
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					return err
				}
			}

			for _, key := range keys {
				fmt.Print(output[key])
			}
			for _, x := range pf {
				fmt.Print(x)
			}
			return nil
		},
	}).Execute(); err != nil {
		panic(err)
	}
}
