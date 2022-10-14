package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/zmb3/spotify/v2"
)

// Source
type SpotifySource struct {
	client             *spotify.Client
	ctx                context.Context
	old_volume         int
	onPlaybackFinished func()
	dbusConn           *dbus.Conn
}

func (c *SpotifySource) register(onPBFinished func()) {
	c.onPlaybackFinished = onPBFinished
}

func (c *SpotifySource) play(music_id Music_ID) {
	uri := music_id.spotify()
	opt := spotify.PlayOptions{
		URIs: []spotify.URI{uri},
	}

	ret, err := json.MarshalIndent(opt, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(ret))

	fmt.Printf("URI: %v\n", opt)
	if err := c.client.PlayOpt(c.ctx, &opt); err != nil {
		panic(err)
	}

	go c.waitForEnd()
}
func (c *SpotifySource) pause() {
	c.client.Pause(c.ctx)
}
func (c *SpotifySource) stop() {
	c.client.Pause(c.ctx)
}
func (c *SpotifySource) skip() {
	c.client.Next(c.ctx)
}
func (c *SpotifySource) resume() {
	c.client.Play(c.ctx)
}
func (c *SpotifySource) forward(ammount int) {
	state, err := c.client.PlayerCurrentlyPlaying(c.ctx)
	if err != nil {
		return
	}

	c.client.Seek(c.ctx, state.Progress+ammount)
}
func (c *SpotifySource) reverse(ammount int) {
	c.forward(-ammount)
}
func (c *SpotifySource) set_volume(vol int) {
	c.client.Volume(c.ctx, vol)
}
func (c *SpotifySource) get_volume() int {
	state, err := c.client.PlayerState(c.ctx)

	//TODO: handle error
	if err != nil {
		return 50
	}

	return state.Device.Volume
}
func (c *SpotifySource) mute() {
	if c.get_volume() != 0 {
		c.old_volume = c.get_volume()
		c.set_volume(0)
	} else {
		c.set_volume(c.old_volume)
	}
}

func (c *SpotifySource) waitForEnd() {
	const WAIT_PERC = 0.95
	const WAIT_CUTOFF = 3 * 1000000 // 3 seconds

	obj := c.dbusConn.Object("org.mpris.MediaPlayer2.spotify", "/org/mpris/MediaPlayer2")

	for {
		//Get current position
		var pos uint64
		err := obj.Call("org.freedesktop.DBus.Properties.Get", 0, "org.mpris.MediaPlayer2.Player", "Position").Store(&pos)
		if err != nil {
			panic(err)
		}

		//Get duration
		var metadata map[string]dbus.Variant
		err = obj.Call("org.freedesktop.DBus.Properties.Get", 0, "org.mpris.MediaPlayer2.Player", "Metadata").Store(&metadata)
		if err != nil {
			panic(err)
		}

		dur := metadata["mpris:length"].Value().(uint64)

		rem := dur - pos

		fmt.Printf("Pos: %v/%v (%v)\n", pos, dur, rem)

		//Wait
		if pos == 0 {
			fmt.Println("FINISHED")
			c.onPlaybackFinished()
			return
		} else if rem < WAIT_CUTOFF && rem > 300000 {
			//Check every 0.1 seconds
			time.Sleep(100 * time.Millisecond)
		} else if rem < 30 {
			fmt.Println("FINISHED")
			c.onPlaybackFinished()
			return
		} else {
			time.Sleep(time.Duration(float64(rem)*WAIT_PERC) * time.Microsecond)
		}

	}
}
