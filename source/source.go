package source

import entry "muzsikusch/queue/entry"

type Source interface {
	Play(entry.MusicID) error
	Pause() error
	Stop() error
	Skip() error
	Resume() error
	Forward(int) error
	Reverse(int) error
	SetVolume(int) error
	GetVolume() (int, error)
	Mute() error
	Register(func())
	Search(string) []entry.MusicID
	GetTimePos() (float32, error)
	BelongsToThis(string) (bool, entry.MusicID) //Checks whether the search query is a valid ID that describes a track from there specifically, if yes, returns with a function that turns the query into a MusicID this was migrated from muscID.go
}
