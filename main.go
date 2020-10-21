package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
)

const (
	EXIT_SUCCESS       = 0
	EXIT_FAILURE       = 1
	EXIT_USAGE_FAILURE = 2 // Same value used by go's flag library, unfortunatey not exported as a constant: https://github.com/golang/go/blob/release-branch.go1.15/src/flag/flag.go#L985
)

func main() {
	dataStoreInputFilename, dataStoreOutputFilename, changesFilename := GetCommandLineArgsOrExit()

	dataStore, err := NewDataStoreFromFile(dataStoreInputFilename)
	if err != nil {
		fmt.Printf("Problem reading the data store file ('%v'): %v\n", dataStoreInputFilename, err)
		os.Exit(EXIT_FAILURE)
	}

	commands, err := LoadChangesFile(changesFilename)
	if err != nil {
		fmt.Printf("Problem reading the changes file ('%v'): %v\n", changesFilename, err)
		os.Exit(EXIT_FAILURE)
	}

	processor := NewCommandProcessor(commands, dataStore)
	if err := processor.ProcessAll(); err != nil {
		fmt.Printf("Problem with executing changes: %v\nList of problems:\n%v\n", err, processor.Errors())
		os.Exit(EXIT_FAILURE)
	}

	if err := dataStore.WriteToFile(dataStoreOutputFilename); err != nil {
		fmt.Printf("Problem writing data store to file ('%v'): %v\n", dataStoreOutputFilename, err)
		os.Exit(EXIT_FAILURE)
	}
}

/* GetCommandLineArgsOrExit returns the user-supplied command line arguments. If any of the three
required arguments are missing, a basic usage text will be shown to the user and the program terminates
with the appropriate exit code. */
func GetCommandLineArgsOrExit() (dataStoreInputFilename string, dataStoreOutputFilename string, changesFilename string) {
	flag.StringVar(&dataStoreInputFilename, "in", "", "File of data store to read.")
	flag.StringVar(&dataStoreOutputFilename, "out", "", "File where results of changes will be written.")
	flag.StringVar(&changesFilename, "changes", "", "File containing the list of changes to apply.")

	flag.Parse()

	if dataStoreInputFilename == "" || dataStoreOutputFilename == "" || changesFilename == "" {
		flag.Usage()
		os.Exit(EXIT_USAGE_FAILURE)
	}

	return
}

/* LoadChangesFile reads the "changes" file and returns a slice of commands. Each command is itself
a slice where each part of the command is an element of the slice.
Example: The command "command-name,1,2" would become the slice []string{"command-name", "1", "2'"} */
func LoadChangesFile(filename string) ([][]string, error) {
	inputFile, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("Couldn't open changes file: %v\n", err)
	}

	inputCsv := csv.NewReader(inputFile)
	inputCsv.Comment = '#'
	inputCsv.FieldsPerRecord = -1 // Disable auto-checking of number of fields per line

	commands, err := inputCsv.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("Problem reading CSV file: %v", err)
	}

	return commands, nil
}
