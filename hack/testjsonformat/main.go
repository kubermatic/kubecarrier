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
	"sort"
	"strings"
	"time"
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
	reader := bufio.NewReader(os.Stdin)
	keys := make([]string, 0)
	output := make(map[string]*strings.Builder, 0)
	fail := make([]string, 0)
	pass := make([]string, 0)

	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			ev := &TestEvent{}
			if err := json.Unmarshal(line, ev); err != nil {
				panic(err)
			}

			if ev.Action == "output" {
				if strings.HasPrefix(ev.Output, "    ") {
					if !strings.HasPrefix(ev.Output, "    ") {
						continue
					}
					if strings.Count(ev.Output, ":") <= 1 {
						continue
					}
					candidate := strings.SplitN(strings.TrimSpace(ev.Output), ":", 2)[0]
					if _, ok := output[candidate]; !ok {
						continue
					}
					ev.Test = candidate
				}
			}

			if _, ok := output[ev.Test]; !ok {
				output[ev.Test] = new(strings.Builder)
				keys = append(keys, ev.Test)
			}
			switch ev.Action {
			case "output":
				output[ev.Test].WriteString(ev.Output)
			case "fail":
				fail = append(fail, ev.Test)
			case "pass":
				pass = append(pass, ev.Test+" "+ev.Output)
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			panic(err)
		}
	}

	sort.Strings(keys)
	for _, key := range keys {
		fmt.Println("TEST", key)
		fmt.Println(output[key])
	}
	for _, p := range pass {
		fmt.Println(p)
	}
	if len(fail) > 0 {
		for _, f := range fail {
			fmt.Println("FAILED", f)
		}
		os.Exit(1)
	}
}
