package main

type Source interface {
	Play(MusicID) error
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
	Search(string) MusicID
}
