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
	shortText := text
	if len(text) > 25 {
		shortText = shortText[:25] + "…"
	}

	return strings.ReplaceAll(shortText, "\n", "⋅")
}
