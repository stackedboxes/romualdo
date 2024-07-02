/******************************************************************************\
* The Romualdo Language                                                        *
*                                                                              *
* Copyright 2020-2024 Leandro Motta Barros                                     *
* Licensed under the MIT license (see LICENSE.txt for details)                 *
\******************************************************************************/

package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/stackedboxes/romualdo/pkg/bytecode"
	"github.com/stackedboxes/romualdo/pkg/vm"
)

var devDisassembleCmd = &cobra.Command{
	Use:   "disassemble <ras-file>",
	Short: "Disassemble a Romualdo compiled storyworld",
	Long:  `Disassemble a Romualdo compiled storyworld.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		csw, di, err := vm.LoadCompiledStoryworldBinaries(args[0], false)
		reportAndExitOnError(err)

		// Basic info
		fmt.Printf("Disassembling %s\n", args[0])
		fmt.Printf("Total %v constants, %v chunks\n", len(csw.Constants), len(csw.Chunks))
		fmt.Printf("Initial chunk: %v %v\n", csw.InitialChunk, chunkDebugInfo(csw, di, csw.InitialChunk))

		// Chunks summary
		fmt.Println("\nChunks summary:")
		for i, c := range csw.Chunks {
			chunkDI := chunkDebugInfo(csw, di, i)
			fmt.Printf("    %5d: %5d bytes long %v\n", i, len(c.Code), chunkDI)
		}

		// Constants
		if flagDevDisassembleConstants || flagDevDisassembleAll {
			fmt.Println("\nConstants:")
			for i, c := range csw.Constants {
				fmt.Printf("    %5d: %v\n", i, c)
			}
		}

		// Full disassembly of requested Procedures
		if len(*flagDevDisassembleProcs) == 0 && !flagDevDisassembleAll {
			return
		}

		for i, c := range csw.Chunks {
			if !shouldDisassembleThisChunk(csw, di, i) {
				continue
			}
			fmt.Printf("\nDisassembly of Chunk %v %v:\n", i, chunkDebugInfo(csw, di, i))
			csw.DisassembleChunk(c, os.Stdout, di, i)
		}

		reportAndExit(nil)
	},
}

// chunkDebugInfo returns a string with debug information about the chunk at
// index idx. The provided di can be nil, in which case an empty string is
// returned.
func chunkDebugInfo(csw *bytecode.CompiledStoryworld, di *bytecode.DebugInfo, idx int) string {
	if di == nil {
		return ""
	}
	return fmt.Sprintf("[%v, %v]", di.ChunksNames[idx], di.ChunksSourceFiles[idx])
}

// shouldDisassembleThisChunk returns true if the chunk at index idx should be
// disassembled. The provided di can be nil, in which case only Chunk indices
// (i.e., no Chink names) will be recognized.
func shouldDisassembleThisChunk(csw *bytecode.CompiledStoryworld, di *bytecode.DebugInfo, idx int) bool {
	if flagDevDisassembleAll {
		return true
	}
	for _, p := range *flagDevDisassembleProcs {
		if strconv.Itoa(idx) == p || (di != nil && p == di.ChunksNames[idx]) {
			return true
		}
	}
	return false
}

// flagDevDisassembleAll is the value of the --all flag of the `dev disassemble`
// command.
var flagDevDisassembleAll bool

// flagDevDisassembleConstants is the value of the --constants flag of the `dev
// disassemble` command.
var flagDevDisassembleConstants bool

// flagDevDisassembleProcs is the value of the --proc flag of the `dev
// disassemble` command.
var flagDevDisassembleProcs *[]string

func init() {
	devDisassembleCmd.Flags().BoolVarP(&flagDevDisassembleAll, "all", "a",
		false, "Disassemble everything in the compiled Storyworld")

	devDisassembleCmd.Flags().BoolVarP(&flagDevDisassembleConstants, "constants", "c",
		false, "List all constants in the compiled Storyworld")

	flagDevDisassembleProcs = devDisassembleCmd.Flags().StringArrayP("proc", "p",
		[]string{}, "Procedures to disassemble (name or index, can be specified multiple times)")
}
