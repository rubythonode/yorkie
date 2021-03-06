/*
 * Copyright 2020 The Yorkie Authors. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package splay_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yorkie-team/yorkie/pkg/splay"
)

type stringValue struct {
	content string
}

func newSplayNode(content string) *splay.Node {
	return splay.NewNode(&stringValue{
		content: content,
	})
}

func (v *stringValue) Len() int {
	return len(v.content)
}

func (v *stringValue) String() string {
	return v.content
}

func TestSplayTree(t *testing.T) {
	t.Run("insert and splay test", func(t *testing.T) {
		tree := splay.NewTree()

		nodeA := tree.Insert(newSplayNode("A2"))
		assert.Equal(t, "[2,2]A2", tree.AnnotatedString())
		nodeB := tree.Insert(newSplayNode("B23"))
		assert.Equal(t, "[2,2]A2[5,3]B23", tree.AnnotatedString())
		nodeC := tree.Insert(newSplayNode("C234"))
		assert.Equal(t, "[2,2]A2[5,3]B23[9,4]C234", tree.AnnotatedString())
		nodeD := tree.Insert(newSplayNode("D2345"))
		assert.Equal(t, "[2,2]A2[5,3]B23[9,4]C234[14,5]D2345", tree.AnnotatedString())

		tree.Splay(nodeB)
		assert.Equal(t, "[2,2]A2[14,3]B23[9,4]C234[5,5]D2345", tree.AnnotatedString())

		assert.Equal(t, tree.IndexOf(nodeA), 0)
		assert.Equal(t, tree.IndexOf(nodeB), 2)
		assert.Equal(t, tree.IndexOf(nodeC), 5)
		assert.Equal(t, tree.IndexOf(nodeD), 9)
	})
}
