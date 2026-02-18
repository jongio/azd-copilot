// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package squad

import "embed"

//go:embed templates/*.md
var embeddedTemplates embed.FS

// readTemplate reads an embedded template file by name.
func readTemplate(name string) []byte {
	data, err := embeddedTemplates.ReadFile("templates/" + name)
	if err != nil {
		// Templates are embedded at compile time; a missing template is a build error.
		panic("squad: missing embedded template: " + name)
	}
	return data
}
