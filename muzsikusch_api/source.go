package main

type Source interface {
	play(MusicID) error
	pause() error
	stop() error
	skip() error
	resume() error
	forward(int) error
	reverse(int) error
	setVolume(int) error
	getVolume() (int, error)
	mute() error
}
