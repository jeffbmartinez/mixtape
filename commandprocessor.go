package main

import (
	"fmt"
)

const ADD_PLAYLIST = "add-playlist"
const ADD_SONG_TO_PLAYLIST = "add-song-to-playlist"
const REMOVE_PLAYLIST = "rm-playlist"

type CommandProcessor struct {
	Commands  [][]string
	DataStore *DataStore

	errors []error
}

func NewCommandProcessor(commands [][]string, dataStore *DataStore) CommandProcessor {
	return CommandProcessor{
		Commands:  commands,
		DataStore: dataStore,
		errors:    []error{},
	}
}

/* ProcessAll processes the commands passed to it.
ProcessAll does a "best attempt" at executing as many commands as possible.
Any errors encountered will be returned in a list after it completes the entire list.
This means if a later command depends on a previous one that has failed, it will also fail.
Any errors encountered during ProcessAll can be grabbed later by calling `Errors()` */
func (cp *CommandProcessor) ProcessAll() error {
	cp.errors = []error{} // reset errors from any previous run

	for _, command := range cp.Commands {
		if err := cp.ProcessCommand(command); err != nil {
			cp.errors = append(cp.errors, err)
		}
	}

	if len(cp.errors) != 0 {
		return fmt.Errorf("At least one error was encountered. See CommandProcessor.Errors()")
	}

	return nil
}

func (cp *CommandProcessor) ProcessCommand(command []string) error {
	if len(command) == 0 {
		return fmt.Errorf("Can't process empty command")
	}

	commandHandlers := map[string]func([]string) error{
		ADD_PLAYLIST:         cp.addNewPlaylist,
		ADD_SONG_TO_PLAYLIST: cp.addSongToPlaylist,
		REMOVE_PLAYLIST:      cp.removePlaylist,
	}

	baseCommand := command[0]
	commandHandler, exists := commandHandlers[baseCommand]
	if !exists {
		return fmt.Errorf("Unrecognized command: `%v`\n", baseCommand)
	}

	if err := commandHandler(command); err != nil {
		return fmt.Errorf("Problem with `%v`: %v\n", command, err)
	}

	return nil
}

func (cp CommandProcessor) Errors() []error {
	return cp.errors
}

// Command format: []string{ADD_PLAYLIST, "playlist-id", "song-id-1", "song-id-2", ..., "song-id-N"}
func (cp *CommandProcessor) addNewPlaylist(command []string) error {
	if len(command) < 3 {
		return fmt.Errorf("Incorrect number of arguments for `%v`", command[0])
	}

	userId := command[1]
	songIds := command[2:]

	if _, err := cp.DataStore.AddNewPlaylist(userId, songIds); err != nil {
		return err
	}

	return nil
}

// Command format: []string{REMOVE_PLAYLIST, "playlist-id"}
func (cp *CommandProcessor) removePlaylist(command []string) error {
	if len(command) != 2 {
		return fmt.Errorf("Incorrect number of arguments for `%v` command", command[0])
	}

	playlistId := command[1]

	if _, err := cp.DataStore.RemovePlaylist(playlistId); err != nil {
		return err
	}

	return nil
}

// Command format: []string{ADD_SONG_TO_PLAYLIST, "playlist-id", "song-id"}
func (cp *CommandProcessor) addSongToPlaylist(command []string) error {
	if len(command) != 3 {
		return fmt.Errorf("Incorrect number of arguments for `%v` command", command[0])
	}

	playlistId := command[1]
	songId := command[2]

	return cp.DataStore.AddSongToPlaylist(playlistId, songId)
}
