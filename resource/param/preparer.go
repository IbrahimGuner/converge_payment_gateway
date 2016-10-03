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

package param

import (
	"errors"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
)

// Preparer for params
//
// Param controls the flow of values through `module` calls. You can use the
// `{{param "name"}}` template call anywhere you need the value of a param
// inside the current module.
type Preparer struct {
	// Default is an optional field that provides a default value if none is
	// provided to this parameter. If this field is not set, this param will be
	// treated as required.
	Default interface{} `hcl:"default"`
}

// Prepare a new task
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	if val, present := render.Value(); present {
		return &Param{Val: val}, nil
	}

	if p.Default == nil {
		return nil, errors.New("param is required")
	}

	return &Param{Val: p.Default}, nil
}

func init() {
	registry.Register("param", (*Preparer)(nil), (*Param)(nil))
}
