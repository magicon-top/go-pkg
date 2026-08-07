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

package form

import "testing"

// TestFieldMapSkipsEmptyImageValue verifies empty CSV image fields do not create image boxes.
func TestFieldMapSkipsEmptyImageValue(t *testing.T) {
	fieldNames := []string{"@img(page:1, pos:40 350, w:290, h:200)"}
	formRecord := []string{""}

	_, images, _, err := FieldMap(fieldNames, formRecord)
	if err != nil {
		t.Fatal(err)
	}
	if len(images) != 0 {
		t.Fatalf("image map size = %d, want 0", len(images))
	}
}
