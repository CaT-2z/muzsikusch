package source

import "muzsikusch/queue"

type TitleResolver interface {
	ResolveTitle(*queue.MusicID) (string, error)
}
