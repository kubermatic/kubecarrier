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

	"github.com/gernest/wow"
)

// AttachSpinnerTo attaches a spinner to a function with the message given.
// For example if you attach a spinner with message "This is a spinner",
// then, during the execution time of the function, the output will be:
// [spinner] This is a spinner...
// And if there is no error returned by the function, the output after the function executed will be:
// ✔ This is a spinner
// If function returns an error, then the output will be:
// ✖ This is a spinner
func AttachSpinnerTo(spinner *wow.Wow, msg string, f func() error) error {
	spinner.Text(fmt.Sprintf(" %s...", msg))
	spinner.Start()
	if err := f(); err != nil {
		spinner.Text(msg + " " + wow.ERROR.String())
		spinner.Persist()
		return fmt.Errorf("%s: %w", msg, err)
	}
	spinner.Text(msg + " " + wow.SUCCESS.String())
	spinner.Persist()
	return nil
}
