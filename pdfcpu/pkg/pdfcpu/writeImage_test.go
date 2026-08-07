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
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// TestValidatePDFImageDimensionsRejectsPixelLimit verifies pixel limits are enforced.
func TestValidatePDFImageDimensionsRejectsPixelLimit(t *testing.T) {
	conf := model.NewDefaultConfiguration()
	conf.Limits.MaxImagePixels = 9
	conf.Limits.MaxImageBytes = 1 << 20

	xRefTable := &model.XRefTable{Conf: conf}
	err := validatePDFImageDimensions(xRefTable, 5, 2, 3, 8, 1)
	if err == nil || !strings.Contains(err.Error(), "pixel count") {
		t.Fatalf("got %v, want pixel count limit error", err)
	}
}

// TestValidatePDFImageDimensionsRejectsByteLimit verifies byte limits are enforced.
func TestValidatePDFImageDimensionsRejectsByteLimit(t *testing.T) {
	conf := model.NewDefaultConfiguration()
	conf.Limits.MaxImagePixels = 100
	conf.Limits.MaxImageBytes = 8

	xRefTable := &model.XRefTable{Conf: conf}
	err := validatePDFImageDimensions(xRefTable, 2, 2, 3, 8, 1)
	if err == nil || !strings.Contains(err.Error(), "byte size") {
		t.Fatalf("got %v, want byte size limit error", err)
	}
}
