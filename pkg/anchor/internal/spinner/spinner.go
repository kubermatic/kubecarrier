/*
Copyright 2019 The KubeCarrier Authors.

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

package spinner

import (
	"fmt"
	"time"

	"github.com/gernest/wow"
	"github.com/gernest/wow/spin"
)

const (
	succeed = "✔ "
	failed  = "✖ "
)

// AttachSpinnerTo attaches a spinner to a function with the message given.
// For example if you attach a spinner with message "This is a spinner",
// then, during the execution time of the function, the output will be:
// [spinner] This is a spinner...
// And if there is no error returned by the function, the output after the function executed will be:
// ✔ This is a spinner
// If function returns an error, then the output will be:
// ✖ This is a spinner
func AttachSpinnerTo(spinner *wow.Wow, startTime time.Time, msg string, f func() error) error {
	spinner.Text(fmt.Sprintf(" %s...", msg))
	spinner.Start()
	if err := f(); err != nil {
		spinner.PersistWith(spin.Spinner{Frames: []string{fmt.Sprintf("%4.2fs %s ", float64(time.Since(startTime))/float64(time.Second), failed)}}, msg)
		return err
	}
	spinner.PersistWith(spin.Spinner{Frames: []string{fmt.Sprintf("%4.2f %s ", float64(time.Since(startTime))/float64(time.Second), succeed)}}, msg)
	return nil
}
