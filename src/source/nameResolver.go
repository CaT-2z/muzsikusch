package source

import (
	"muzsikusch/src/queue/entry"
)

type TitleResolver interface {
	ResolveTitle(*entry.MusicID) (string, error)
}
