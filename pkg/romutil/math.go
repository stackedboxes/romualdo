/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2024 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package romutil

// Abs returns the absolute value of i. I hear this is too simple to be on the
// standard library. *sigh*
func Abs(i int32) int32 {
	if i < 0 {
		return -i
	}
	return i
}
