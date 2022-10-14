package main

type Source interface {
	play(Music_ID)
	pause()
	stop()
	skip()
	resume()
	forward(int)
	reverse(int)
	setVolume(int)
	getVolume() int
	mute()
}
