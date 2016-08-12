// Copyright © 2016 Asteris, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plan

import "github.com/asteris-llc/converge/resource"

// Result is the result of planning execution
type Result struct {
	Task   resource.Task
	Status resource.TaskStatus
	Err    error
}

// Fields returns the fields that will change based on this result
func (r *Result) Fields() map[string][2]string {
	diffOutput := make(map[string][2]string)
	for key, diff := range r.Status.Diffs() {
		if diff.Changes() {
			diffOutput[key] = [2]string{diff.Original(), diff.Current()}
		}
	}
	return diffOutput
}

// HasChanges indicates if this result will change
func (r *Result) HasChanges() bool { return r.Status.Changes() }

// Error returns the error assigned to this Result, if any
func (r *Result) Error() error {
	return r.Err
}
