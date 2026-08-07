/*
	Copyright 2026 The pdfcpu Authors.

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

package pdfcpu

import (
	"errors"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func nestedBookmarks(depth int) []Bookmark {
	bms := []Bookmark{{Title: "bookmark", PageFrom: 1}}
	if depth > 0 {
		bms[0].Kids = nestedBookmarks(depth - 1)
	}
	return bms
}

// TestBookmarkListRejectsRecursionDepth verifies bookmark listing respects recursion limits.
func TestBookmarkListRejectsRecursionDepth(t *testing.T) {
	_, err := bookmarkList(nestedBookmarks(2), 0, 1)
	if !errors.Is(err, model.ErrMaxRecursionDepthExceeded) {
		t.Fatalf("got %v, want ErrMaxRecursionDepthExceeded", err)
	}
}

// TestCreateOutlineItemDictRejectsRecursionDepth verifies outline creation respects recursion limits.
func TestCreateOutlineItemDictRejectsRecursionDepth(t *testing.T) {
	_, _, _, _, err := createOutlineItemDictDepth(nil, nestedBookmarks(0), nil, nil, model.DefaultResourceLimits().MaxRecursionDepth+1)
	if !errors.Is(err, model.ErrMaxRecursionDepthExceeded) {
		t.Fatalf("got %v, want ErrMaxRecursionDepthExceeded", err)
	}
}

func cyclicBookmarkContext(t *testing.T) *model.Context {
	t.Helper()

	ctx, err := model.NewContext(strings.NewReader(""), model.NewDefaultConfiguration())
	if err != nil {
		t.Fatal(err)
	}

	dest := types.Array{types.Integer(1)}
	d := types.Dict{
		"Title": types.StringLiteral("bookmark"),
		"Dest":  dest,
		"Next":  *types.NewIndirectRef(1, 0),
	}
	if _, err := ctx.IndRefForObject(1, d); err != nil {
		t.Fatal(err)
	}
	return ctx
}

// TestBookmarksForOutlineItemRejectsCycle verifies outline traversal rejects cycles.
func TestBookmarksForOutlineItemRejectsCycle(t *testing.T) {
	ctx := cyclicBookmarkContext(t)

	_, err := BookmarksForOutlineItem(ctx, types.NewIndirectRef(1, 0), nil)
	if !errors.Is(err, errCircularBookmarks) {
		t.Fatalf("got %v, want errCircularBookmarks", err)
	}
}

// TestRemoveNamedDestsRejectsCycle verifies named destination removal rejects cycles.
func TestRemoveNamedDestsRejectsCycle(t *testing.T) {
	ctx := cyclicBookmarkContext(t)

	err := removeNamedDests(ctx, types.NewIndirectRef(1, 0), 0, map[int]bool{})
	if !errors.Is(err, errCircularBookmarks) {
		t.Fatalf("got %v, want errCircularBookmarks", err)
	}
}
