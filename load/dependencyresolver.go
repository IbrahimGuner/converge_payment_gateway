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

package load

import (
	"context"
	"fmt"
	"io/ioutil"
	"sort"
	"sync"
	"text/template"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/parse"
	"github.com/asteris-llc/converge/render/extensions"
	"github.com/asteris-llc/converge/render/preprocessor"
)

type dependencyGenerator func(g *graph.Graph, id string, node *parse.Node) ([]string, error)

// ResolveDependencies examines the strings and depdendencies at each vertex of
// the graph and creates edges to fit them
func ResolveDependencies(ctx context.Context, g *graph.Graph) (*graph.Graph, error) {
	logger := logging.GetLogger(ctx).WithField("function", "ResolveDependencies")
	logger.Debug("resolving dependencies")

	groupLock := new(sync.RWMutex)
	groupMap := make(map[string][]string)
	g, err := g.Transform(ctx, func(meta *node.Node, out *graph.Graph) error {
		if graph.IsRoot(meta.ID) { // skip root
			return nil
		}

		node, ok := meta.Value().(*parse.Node)
		if !ok {
			return fmt.Errorf("ResolveDependencies can only be used on Graphs of *parse.Node. I got %T", meta.Value())
		}

		depGenerators := []dependencyGenerator{getDepends, getParams, getXrefs}

		// we have dependencies from various sources, but they're always IDs, so we
		// can connect them pretty easily
		for _, source := range depGenerators {
			deps, err := source(g, meta.ID, node)
			if err != nil {
				return err
			}
			for _, dep := range deps {
				out.Connect(meta.ID, dep)
			}
		}

		// collect groups information
		group, err := groupName(node)
		if err != nil {
			return fmt.Errorf("failed to retrieve group from node %s", meta.ID)
		}
		if group != "" {
			groupLock.Lock()
			groupMap[group] = append(groupMap[group], meta.ID)
			groupLock.Unlock()
		}

		return nil
	})

	// create dependencies between nodes in each group
	for _, ids := range groupMap {
		// sort ids so that intra-group dependencies are prioritized
		sort.Strings(ids)
		for i, id := range ids {
			if i > 0 {
				from := id
				to := ids[i-1]

				groupDep := func(id string) string {
					pid := graph.ParentID(id)
					if !graph.IsRoot(pid) {
						id = pid
					}
					return id
				}

				if !graph.AreSiblingIDs(from, to) {
					from = groupDep(from)
					to = groupDep(to)
				}

				g.Connect(from, to)
			}
		}
	}

	return g, err
}

func groupName(node *parse.Node) (string, error) {
	group, err := node.GetString("group")
	switch err {
	case parse.ErrNotFound:
		return "", nil
	case nil:
		return group, nil
	default:
		return "", err
	}
}

func getDepends(g *graph.Graph, id string, node *parse.Node) ([]string, error) {
	deps, err := node.GetStringSlice("depends")
	switch err {
	case parse.ErrNotFound:
		return []string{}, nil
	case nil:
		for idx, dep := range deps {
			if ancestor, ok := getNearestAncestor(g, id, dep); ok {
				deps[idx] = ancestor
			} else {
				return nil, fmt.Errorf("nonexistent vertices in edges: %s", dep)
			}
		}
		return deps, nil
	default:
		return nil, err
	}
}

func getParams(g *graph.Graph, id string, node *parse.Node) (out []string, err error) {
	var nodeStrings []string
	nodeStrings, err = node.GetStrings()
	if err != nil {
		return nil, err
	}

	type stub struct{}
	language := extensions.MinimalLanguage()
	language.On("param", extensions.RememberCalls(&out, ""))
	language.On("paramList", extensions.RememberCalls(&out, []interface{}(nil)))
	language.On("paramMap", extensions.RememberCalls(&out, map[string]interface{}(nil)))

	for _, s := range nodeStrings {
		useless := stub{}
		tmpl, tmplErr := template.New("DependencyTemplate").Funcs(language.Funcs).Parse(s)
		if tmplErr != nil {
			return out, tmplErr
		}
		tmpl.Execute(ioutil.Discard, &useless)
	}
	for idx, val := range out {
		ancestor, found := getNearestAncestor(g, id, "param."+val)
		if !found {
			return out, fmt.Errorf("unknown parameter: param.%s", val)
		}
		out[idx] = ancestor
	}
	return out, err
}

func getXrefs(g *graph.Graph, id string, node *parse.Node) (out []string, err error) {
	var nodeStrings []string
	var calls []string
	nodeRefs := make(map[string]struct{})
	nodeStrings, err = node.GetStrings()
	if err != nil {
		return nil, err
	}
	language := extensions.MinimalLanguage()
	language.On(extensions.RefFuncName, extensions.RememberCalls(&calls, 0))
	for _, s := range nodeStrings {
		tmpl, tmplErr := template.New("DependencyTemplate").Funcs(language.Funcs).Parse(s)
		if tmplErr != nil {
			return out, tmplErr
		}
		tmpl.Execute(ioutil.Discard, &struct{}{})
	}
	for _, call := range calls {
		vertex, _, found := preprocessor.VertexSplitTraverse(g, call, id, preprocessor.TraverseUntilModule, make(map[string]struct{}))
		if !found {
			return []string{}, fmt.Errorf("dependency generator: unresolvable call to %s", call)
		}
		if _, ok := nodeRefs[vertex]; !ok {
			nodeRefs[vertex] = struct{}{}
			out = append(out, vertex)
			if peerVertex, ok := getPeerVertex(id, vertex); ok {
				out = append(out, peerVertex)
			}
		}
	}
	return out, err
}

func getPeerVertex(src, dst string) (string, bool) {
	if dst == "." || graph.IsRoot(dst) {
		return "", false
	}
	if graph.AreSiblingIDs(src, dst) {
		return dst, true
	}
	return getPeerVertex(src, graph.ParentID(dst))
}

func getNearestAncestor(g *graph.Graph, id, node string) (string, bool) {
	if graph.IsRoot(id) || id == "" || id == "." {
		return "", false
	}

	siblingID := graph.SiblingID(id, node)

	valMeta, ok := g.Get(siblingID)
	if !ok {
		return getNearestAncestor(g, graph.ParentID(id), node)
	}
	_, ok = valMeta.Value().(*parse.Node)
	if !ok {
		return "", false
	}
	return siblingID, true
}
