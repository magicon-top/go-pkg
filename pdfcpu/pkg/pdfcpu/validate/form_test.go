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
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// TestValidateFormFieldDictRejectsCycle verifies form field validation rejects cycles.
func TestValidateFormFieldDictRejectsCycle(t *testing.T) {
	ctx, err := model.NewContext(strings.NewReader(""), model.NewDefaultConfiguration())
	if err != nil {
		t.Fatal(err)
	}

	ir := *types.NewIndirectRef(1, 0)
	if _, err := ctx.IndRefForObject(1, types.Dict{
		"Kids": types.Array{ir},
	}); err != nil {
		t.Fatal(err)
	}

	err = validateFormFieldDict(ctx.XRefTable, ir, nil, false)
	if !errors.Is(err, model.ErrFormFieldCycle) {
		t.Fatalf("got %v, want ErrFormFieldCycle", err)
	}
}
