package main

type TitleResolver interface {
	ResolveTitle(*MusicID) (string, error)
}
