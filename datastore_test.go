package main

import (
	"testing"
)

func TestNewDataStoreFromFile(t *testing.T) {
	dataStore, err := NewDataStoreFromFile("testtape.json")
	if err != nil {
		t.Error("Error during data store file read: ", err)
	}

	if len(dataStore.Users) != 2 {
		t.Error("Expected to read in 2 users from data store")
	}

	if len(dataStore.Playlists) != 3 {
		t.Error("Expected to read 3 playlists from data store")
	}

	if dataStore.Songs[2].Artist != "The Weeknd" {
		t.Error("Didn't find expected artist name")
	}

	// Lookup table checks

	if dataStore.userMap["1"] != 0 {
		t.Error("Users lookup table not properly initialized")
	}

	if dataStore.songMap["2"] != 1 {
		t.Error("Songs lookup table not properly initialized")
	}

	if dataStore.playlistMap["3"] != 2 {
		t.Error("Playlsits lookup table not properly initialized")
	}

	if dataStore.nextPlaylistId != 4 {
		t.Error("nextPlaylistId not properly initialized")
	}
}

func TestRemovePlaylist(t *testing.T) {
	dataStore, err := NewDataStoreFromFile("testtape.json")
	if err != nil {
		t.Error("Error during data store file read: ", err)
	}

	if removed, err := dataStore.RemovePlaylist("1"); err != nil || !removed {
		t.Error("Failed to remove playlist '1'")
	}
	if len(dataStore.Playlists) != 2 {
		t.Error("Didn't actually remove the playlist")
	}
	if removed, err := dataStore.RemovePlaylist("1"); err != nil || removed {
		t.Error("Should be ok to call remove playlist '1' twice but method should return `false` the second time")
	}
	if dataStore.Playlists[0].Id != "3" || dataStore.Playlists[1].Id != "2" {
		t.Error("Removed the wrong playlist")
	}
}

func TestAddNewPlaylist(t *testing.T) {
	dataStore, err := NewDataStoreFromFile("testtape.json")
	if err != nil {
		t.Error("Error during data store file read: ", err)
	}

	if _, err := dataStore.AddNewPlaylist("bad-id", []string{"1", "2"}); err == nil {
		t.Error("AddNewPlaylist not failing when user id doesn't exist in store")
	}
	if _, err := dataStore.AddNewPlaylist("1", []string{"bad-id", "2"}); err == nil {
		t.Error("AddNewPlaylist not failing when song id doesn't exist in store")
	}
	if _, err := dataStore.AddNewPlaylist("1", []string{}); err == nil {
		t.Error("AddNewPlaylist not failing when the song id list is empty")
	}

	newPlaylistId, err := dataStore.AddNewPlaylist("1", []string{"1", "2"})
	if err != nil || newPlaylistId != "4" {
		t.Error("AddNewPlaylist not adding playlist as expected")
	}
}

func TestAddSongToPlaylist(t *testing.T) {
	dataStore, err := NewDataStoreFromFile("testtape.json")
	if err != nil {
		t.Error("Error during data store file read: ", err)
	}

	if err := dataStore.AddSongToPlaylist("bad-id", "1"); err == nil {
		t.Error("AddSongToPlaylist should fail when song id is invalid")
	}
	if err := dataStore.AddSongToPlaylist("1", "bad-id"); err == nil {
		t.Error("AddSongToPlaylist should fail when playlist id is invalid")
	}

	if err := dataStore.AddSongToPlaylist("1", "1"); err != nil {
		t.Error("Problem adding song to playlist")
	}
	if err := dataStore.AddSongToPlaylist("1", "1"); err != nil {
		t.Error("Problem adding song to playlist")
	}
	if err := dataStore.AddSongToPlaylist("1", "1"); err != nil {
		t.Error("Problem adding song to playlist")
	}

	if len(dataStore.Playlists[0].SongIds) != 5 {
		t.Error("Incorrect number of songs were added to playlist")
	}
}
