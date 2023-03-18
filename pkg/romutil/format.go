/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2023 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package romutil

import "strings"

// Formats text for display -- "text" in the sense of a text token (i.e., the
// kind of text we output in a Romualdo storyworld).
func FormatTextForDisplay(text string) string {
	result := strings.ReplaceAll(text, "\n", "⋅")
	if len(result) > 25 {
		return result[:25] + "…"
	}
	return text
}
