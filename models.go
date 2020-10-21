package main

type User struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Playlist struct {
	Id      string   `json:"id"`
	UserId  string   `json:"user_id"`
	SongIds []string `json:"song_ids"`
}

type Song struct {
	Id     string `json:"id"`
	Artist string `json:"artist"`
	Title  string `json:"title"`
}
