# Mixtape

## Usage

This assumes you have [Go working on your system](https://golang.org/doc/install).

>If you're on osx and have `brew`, I *believe* you can do a `brew install go` just to run this quickly

In the directory/folder with the code:

```
go run . -in mixtape.json -changes changes.csv -out out.json
```

OR

```
go build .
./mixtape -in mixtape.json -changes changes.csv -out out.json
```

In the above examples:
* `mixtape.json` is the input file containing the initial data.
* `out.json` is the resulting file with changes applied (It will be overwritten if it exists or created if it does not).
* `changes.csv` contains a list of instructions/commands to be executed on the data. The format is CSV with the first column being a command and the remaining columns used as arguments for that command. Possible commands are:
  * `add-playlist,user-id,song-id-1,song-id-2,...,song-id-n`
  * `add-song-to-playlist,playlist-id,song-id`
  * `rm-playlist,playlist-id`
  * A `#` character at the beginning of a line is ignored as a comment.

## Current Limitations

### Technical Limitations

* The size of the input files is limited by the amount of memory available on the host system. Because it holds the complete contents of the files in memory before processing begins, the system's memory is a major limiting factor.
* Currently only a single processes can safely manipulate the state of the data at a time. Concurrent modifications (additions or removals) would have a real chance of leaving the data in an inconsistent state. For example, process A adds a song to a playlist while process B removes the same playlist.
* `DataStore`'s `nextPlaylistId` attribute will eventually run out of useable IDs because it just keeps incrementing by 1 to create new ones. In particular, on a 32 bit system, Go's `int` will default to a 32-bit int, which could realistically overflow. Using a 64-bit int would fix this for the foreseeable future, but also UUIDs could help, at the expense of eating up a bit more memory.

### Business Limitations

On a system with 4 GB of memory, we can house information for approximately:
* 5 million songs (~ 1GB)
* 1.3 million users (~ 2GB)
* Plus the extra GB for a *huge* changes file.

This seems more than reasonable for a good-sized startup, but it is still ~1% the userbase of Spotify. There is certainly room for vertical scaling by adding more memory, but in the long term a more scalable solution will be useful to a company with bigger visions.

#### Where do those numbers come from?

The approximations include some assumptions based on the sizes of the songs, playlists, and users. The approximations are based on the data supplied in `mixtape.json`.

* ~200 bytes per song.
* ~1600 total bytes per user (including playlists):
  * 100 bytes per user data
  * 10 playlists per user
  * ~150 bytes per playlist

#### Other useful facts

* Apple Music and Spotify each claim to offer about 50 million songs
* Spotify reported 138 million users at the end of 2020 Q2

## Future Scaling Considerations

### Handling a large "changes" file

Currently, the changes file is read into memory completely before processing begins.

The changes file format is CSV, with one command/instruction per line. This means it is possible to read one line (and command) at a time. Effectively, this treats the file as a stream.

This has the added benefit of allowing a constant stream of commands to be read from any file stream, including `stdin`, for example. They could even be piped in or redirected from another command line program which generates the instructions as needed. In other words, a "pipe and filter" architecture.

### Handline a large `mixtape.json` file

Handling a larger `mixtape.json` file requires two additional considerations:

1. Reading the big file: Reading the file line-by-line is problematic because the JSON format won't usually allow for it. Furthermore, it's completely legal to have a *huge* JSON file all one a single line.
1. Memory limits: Even if you could, the current implementation still requires the eventual in-memory storage of all the data. Reading line-by-line doesn't fix anything in that sense.

#### Reading the big file

Instead of line-by-line, the file would be read in token-by-token. This means replacing the usage of `ioutil.ReadFile` and `json.Unmarshal` with a smarter usage of `json.Decoder` and it's `json.Token` method. This implementation would read parseable "chunks" of the JSON file to generate tokens (similar to the line-by-line method for the CSV file, but with JSON tokens) and process those as they were encountered. With this implementation we never read more than a token or two at a time.

While this reduces the memory finger print for *reading* the JSON file, we still have to store it somewhere...

#### Memory Limits

Given that memory is a bottleneck for this system, moving the "live" storage of the data to a distributed storage system would be beneficial.

Key-value storage style NoSQL databases would be good candidates for this data store. Given the use cases required, I would look for something on the simpler side with minimal bells and whistles.

The main consideration is to have a distributed data store which allows indexing data by at least one key (user id, playlist id, and song id).

Redis, for example, is relatively easy to understand and has the minimum built-in types to keep similar data structures in place (lists, hashes, and sorted sets). In addition, it would allow for the possibility of concurrent modification of the data store in the future, since Redis has built-in transaction support.

As a bonus, given that fact that users probably wont be modifying other users' playlists, it's unlikely that the locking mechanisms of the Redis transactions would slow anything down in a way that would effect the end users meaningfully. For example, User A shouldn't be adding a song to User B's playlist while User B is deleting that playlist.

### Price of the "scaled" version

At current pricing on redislabs.com, a 500 GB redis cluster costs $16.31/hour (~$400/day, ~$12000/month). This would buy close to what would be needed to approach the estimates above for a "spotify scale" system.
