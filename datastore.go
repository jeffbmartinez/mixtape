package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

const BASE_10 = 10
const SIZE_64BIT = 64

/* DataStore has two purposes:
1. Serialization and deserialization of the data store.
2. Contains logic for manipulating the data store (adding playlists/songs and removing playlists)

As a reminder, by default, Go serializes *only* public attritutes, which are those that begin with
a capital letter. In this case, only Users, Playlists, and Songs are (de)serialized. The rest are
private and used for the manipulation logic.

DataStore maintains internal look-up tables which serve as indexes for Users, Playlists, and Songs.
*/
type DataStore struct {
	Users     []User     `json:"users"`
	Playlists []Playlist `json:"playlists"`
	Songs     []Song     `json:"songs"`

	userMap     map[string]int
	playlistMap map[string]int
	songMap     map[string]int

	nextPlaylistId int
}

/* NewDataStoreFromFile creates a DataStore object populated with the contents of `inputFilename`.
An error will be returned in the following cases:
- `inputFilename` cannot be read (doesn't exist, lack of permissions, etc).
- The data store has inconsistent data, for example, a playlist refers to a non-existant song or user id. */
func NewDataStoreFromFile(inputFilename string) (*DataStore, error) {
	data, err := ioutil.ReadFile(inputFilename)
	if err != nil {
		return &DataStore{}, err
	}

	var dataStore *DataStore
	if err := json.Unmarshal(data, &dataStore); err != nil {
		return dataStore, err
	}

	if err := dataStore.buildLookupTables(); err != nil {
		return dataStore, err
	}

	return dataStore, nil
}

func (ds *DataStore) buildLookupTables() error {
	// Users lookup table
	ds.userMap = make(map[string]int)
	for i, user := range ds.Users {
		ds.userMap[user.Id] = i
	}

	// Songs lookup table
	ds.songMap = make(map[string]int)
	for i, song := range ds.Songs {
		ds.songMap[song.Id] = i
	}

	// Playlists lookup table
	// This code also take advantages of the required O(n) loop and determines the initial value
	//   for `nextPlaylistId` by adding 1 to the highest id found
	// Also an opportunity to verify that playlists don't refer to non-existant user or song ids
	ds.playlistMap = make(map[string]int)
	var maxPlaylistId = 0
	for i, playlist := range ds.Playlists {
		ds.playlistMap[playlist.Id] = i

		if intId, err := strconv.Atoi(playlist.Id); err != nil {
			return err
		} else if intId > maxPlaylistId {
			maxPlaylistId = intId
		}
	}

	ds.nextPlaylistId = maxPlaylistId + 1

	return nil
}

/* WriteToFile persists the data store to the file specified by `outputFilename`.
See the mid-method constant `OUT_FILE_PERMISSIONS` for the permissions used when creating the file.
An error can be returned if the file cannot be written for any reason. */
func (ds *DataStore) WriteToFile(outputFilename string) error {
	const PREFIX_STRING = ""
	const INDENT_STRING = "  " // Note: That's two spaces, not one
	dataStoreAsJSON, err := json.MarshalIndent(ds, PREFIX_STRING, INDENT_STRING)
	if err != nil {
		return err
	}

	// Standard Unix rwxrwxrwx style permissions, 0644 = user: r+w, group: r, other: r
	const OUT_FILE_PERMISSIONS os.FileMode = 0644
	if err := ioutil.WriteFile(outputFilename, dataStoreAsJSON, OUT_FILE_PERMISSIONS); err != nil {
		return err
	}

	return nil
}

/* RemovePlaylist removes a playlist from the data store. Removing the same playlist ID twice
has no additional effect and is allowed.
RemovePlaylist returns `true` if the playlist was removed and `false` if no action was taken.
RemovePlaylist currently always returns `nil` as the error, it is left in place to match the signature
of the other DataStore manipulation methods.

In order to avoid copying memory over to fill the gap of the removed playlist from the slice (internally
represented as an array by Go), this method simply overwrites the playlist to be removed with the last
playlist in the slice/array. Then the last playlist (which was jus copied into the old "gap") is removed.
This keeps the operation as O(1) time vs O(n).
The trade-off is that the order of the playlists is not maintained. Because of the nature of this program
and the use cases required, this is an acceptable trade-off.
*/
func (ds *DataStore) RemovePlaylist(id string) (bool, error) {
	targetPlaylistIndex, ok := ds.playlistMap[id]
	if !ok {
		return false, nil
	}

	lastIndex := len(ds.Playlists) - 1
	lastPlaylist := ds.Playlists[lastIndex]

	// remove the playlist from the list of playlists
	ds.Playlists[targetPlaylistIndex] = lastPlaylist
	ds.Playlists = ds.Playlists[:lastIndex]

	// Update the playlist map
	delete(ds.playlistMap, id)
	ds.playlistMap[lastPlaylist.Id] = targetPlaylistIndex

	return true, nil
}

/* AddNewPlaylist adds a new playlist to the data store.
Returns the ID of the new playlist.
An error will be returned in the following cases:
- The user ID does not exist.
- At least one sing ID was provided that doesn't exist.
- Playlists without at least one song are not allowed. */
func (ds *DataStore) AddNewPlaylist(userId string, songIds []string) (string, error) {
	if _, exists := ds.userMap[userId]; !exists {
		return "", fmt.Errorf("The user id does not exist")
	}

	if len(songIds) == 0 {
		return "", fmt.Errorf("A playlist must contain at least one song (zero sing IDs were provided)")
	}

	for _, songId := range songIds {
		if _, exists := ds.songMap[songId]; !exists {
			return "", fmt.Errorf("One or more of the song IDs provided is invalid")
		}
	}

	// Make a copy of the song IDs to prevent the caller from accidentally modifying the song
	// IDs after they've been stored.
	songIdsCopy := make([]string, len(songIds))
	copy(songIdsCopy, songIds)

	newPlaylist := Playlist{
		Id:      ds.generatePlaylistId(),
		UserId:  userId,
		SongIds: songIdsCopy,
	}

	ds.Playlists = append(ds.Playlists, newPlaylist)
	ds.playlistMap[newPlaylist.Id] = len(ds.Playlists) - 1

	return newPlaylist.Id, nil
}

/* AddSongToPlaylist adds an existing song id to a playlist. Duplicate songs are allowed.
An error will be returned in the following cases:
- Song ID doesn't exist
- Playlist ID doesn't exist */
func (ds *DataStore) AddSongToPlaylist(playlistId string, songId string) error {
	if _, exists := ds.songMap[songId]; !exists {
		return fmt.Errorf("Song id does not exist")
	}

	playlistIndex, exists := ds.playlistMap[playlistId]
	if !exists {
		return fmt.Errorf("Playlist id does not exist")
	}

	songIds := ds.Playlists[playlistIndex].SongIds
	ds.Playlists[playlistIndex].SongIds = append(songIds, songId)

	return nil
}

func (ds *DataStore) generatePlaylistId() string {
	playlistId := ds.nextPlaylistId
	ds.nextPlaylistId += 1

	return strconv.Itoa(playlistId)
}
