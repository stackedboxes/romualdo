/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2024 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package main

import "github.com/spf13/cobra"

var devCmd = &cobra.Command{
	Use:   "dev <subcommand>",
	Short: "Collection of subcommands for developing Romualdo itself",
	Long: `Collection of subcommands useful for developing Romualdo itself.
If you are not working to improve the 'romualdo' tool, you probably
don't need to look here.`,
}
