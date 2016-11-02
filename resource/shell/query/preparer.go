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

package query

import (
	"fmt"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/shell"
	"golang.org/x/net/context"
)

// Preparer handles querying
type Preparer struct {
	Interpreter string            `hcl:"interpreter"`
	Query       string            `hcl:"query"`
	CheckFlags  []string          `hcl:"check_flags"`
	ExecFlags   []string          `hcl:"exec_flags"`
	Timeout     string            `hcl:"timeout" doc_type:"duration string"`
	Dir         string            `hcl:"dir"`
	Env         map[string]string `hcl:"env"`
}

// Prepare creates a new query type
func (p *Preparer) Prepare(ctx context.Context, render resource.Renderer) (resource.Task, error) {
	shPrep := &shell.Preparer{
		Interpreter: p.Interpreter,
		Check:       p.Query,
		CheckFlags:  p.CheckFlags,
		ExecFlags:   p.ExecFlags,
		Timeout:     p.Timeout,
		Dir:         p.Dir,
		Env:         p.Env,
	}

	task, err := shPrep.Prepare(ctx, render)

	if err != nil {
		return &Query{}, err
	}

	shell, ok := task.(*shell.Shell)
	if !ok {
		return &Query{}, fmt.Errorf("expected *shell.Shell but got %T", task)
	}

	return &Query{Shell: shell}, nil
}

func init() {
	registry.Register("task.query", (*Preparer)(nil), (*Query)(nil))
}
