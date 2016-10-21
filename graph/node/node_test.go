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

package node_test

import (
	"testing"

	"github.com/asteris-llc/converge/graph/node"
	"github.com/stretchr/testify/assert"
)

// BenchmarkWithValue is mostly for testing the amount of allocations performed
// by node.WithValue
func BenchmarkWithValue(b *testing.B) {
	source := node.New("test", 1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source.WithValue(i)
	}
}

// TestWithValue tests the functionality of node.WithValue
func TestWithValue(t *testing.T) {
	t.Run("base", func(t *testing.T) {
		fst := node.New("test", 1)
		snd := fst.WithValue(2)

		assert.Equal(t, snd.Value(), 2)
	})

	t.Run("shadowing", func(t *testing.T) {
		// TODO: when this grows any pointers, they should be tested here too
		t.Run("no b to a", func(t *testing.T) {
			fst := node.New("test", 1)
			snd := fst.WithValue(2)
			snd.ID = "other"

			assert.Equal(t, fst.ID, "test")
		})

		t.Run("no b to a", func(t *testing.T) {
			fst := node.New("test", 1)
			snd := fst.WithValue(2)
			fst.ID = "other"

			assert.Equal(t, snd.ID, "test")
		})
	})
}

// TestWithGroupable tests that group is set when the value is Groupable
func TestWithGroupable(t *testing.T) {
	t.Parallel()

	t.Run("New", func(t *testing.T) {
		n := node.New("test", &aGroupable{group: "somegroup"})
		assert.Equal(t, "somegroup", n.Group)
	})

	t.Run("WithValue", func(t *testing.T) {
		fst := node.New("test", 1)
		assert.Equal(t, "", fst.Group)

		snd := fst.WithValue(&aGroupable{group: "somegroup"})
		assert.Equal(t, "somegroup", snd.Group)
	})
}

type aGroupable struct {
	group string
}

func (a *aGroupable) Group() string { return a.group }
