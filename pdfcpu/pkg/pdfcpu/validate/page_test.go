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

package validate

import (
	"errors"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// TestValidatePagesDictRejectsRecursionDepth verifies page validation respects recursion limits.
func TestValidatePagesDictRejectsRecursionDepth(t *testing.T) {
	ctx, err := model.NewContext(strings.NewReader(""), model.NewDefaultConfiguration())
	if err != nil {
		t.Fatal(err)
	}

	pageCount := 0
	err = validatePagesDictDepth(ctx.XRefTable, nil, 1, false, nil, &pageCount, ctx.XRefTable.MaxRecursionDepth()+1, model.NewPageTreeVisit())
	if !errors.Is(err, model.ErrMaxRecursionDepthExceeded) {
		t.Fatalf("got %v, want ErrMaxRecursionDepthExceeded", err)
	}
}
