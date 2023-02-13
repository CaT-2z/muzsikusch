package source

import entry "muzsikusch/queue/entry"

type TitleResolver interface {
	ResolveTitle(*entry.MusicID) (string, error)
}
