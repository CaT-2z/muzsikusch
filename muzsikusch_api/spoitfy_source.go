package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/zmb3/spotify/v2"
)

// Source
type SpotifySource struct {
	client             *spotify.Client
	playerDevice       spotify.ID
	ctx                context.Context
	old_volume         int
	onPlaybackFinished func()
	dbusConn           *dbus.Conn
	waiterEnder        context.CancelFunc
}

func (c *SpotifySource) register(onPBFinished func()) {
	c.onPlaybackFinished = onPBFinished
}

func (c *SpotifySource) play(music_id MusicID) error {
	log.Printf("Playing %v\n", music_id.SpotifyID)
	uri := music_id.spotify()
	opt := spotify.PlayOptions{
		DeviceID: &c.playerDevice,
		URIs:     []spotify.URI{uri},
	}

	if err := c.client.PlayOpt(c.ctx, &opt); err != nil {
		spotErr, ok := err.(spotify.Error)
		if !ok {
			return err
		}

		if spotErr.Status == 404 {
			//Device not found
			fmt.Println("Discovering devices")
			c.discoverDevices()
			return c.play(music_id)
		}
	}

	//Check if we have a waiting ctx
	if c.waiterEnder != nil {
		c.waiterEnder()
	}

	ctx, waiterEnder := context.WithCancel(c.ctx)
	c.waiterEnder = waiterEnder

	go c.waitForEnd(ctx)
	return nil
}

func (c *SpotifySource) pause() error {
	return c.client.Pause(c.ctx)
}

func (c *SpotifySource) stop() error {
	return c.client.Pause(c.ctx)
}
func (c *SpotifySource) skip() error {
	return c.client.Next(c.ctx)
}
func (c *SpotifySource) resume() error {
	return c.client.Play(c.ctx)
}
func (c *SpotifySource) forward(ammount int) error {
	state, err := c.client.PlayerCurrentlyPlaying(c.ctx)
	if err != nil {
		return err
	}

	return c.client.Seek(c.ctx, state.Progress+ammount)
}
func (c *SpotifySource) reverse(ammount int) error {
	return c.forward(-ammount)
}
func (c *SpotifySource) setVolume(vol int) error {
	return c.client.Volume(c.ctx, vol)
}
func (c *SpotifySource) getVolume() (int, error) {
	state, err := c.client.PlayerState(c.ctx)

	//TODO: handle error
	if err != nil {
		return 0, err
	}

	return state.Device.Volume, nil
}

func (c *SpotifySource) mute() error {
	vol, err := c.getVolume()
	if err != nil {
		return err
	}

	if vol != 0 {
		c.old_volume = vol
		return c.setVolume(0)
	} else {
		return c.setVolume(c.old_volume)
	}
}

func (c *SpotifySource) search(query string) MusicID {
	results, err := c.client.Search(c.ctx, query, spotify.SearchTypeTrack)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found track %v\n", results.Tracks.Tracks[0].Name)
	return MusicID{
		SpotifyID: string(results.Tracks.Tracks[0].URI),
	}
}

func (c *SpotifySource) discoverDevices() {
	devs, err := c.client.PlayerDevices(c.ctx)
	if err != nil {
		panic(err)
	}

	for dev := range devs {
		//TODO: Change this
		if devs[dev].Name == "archertwo" {
			c.playerDevice = devs[dev].ID
			return
		}
	}
}

func (c *SpotifySource) waitForEnd(ctx context.Context) {
	log.Println("Waiter started")
	const WAIT_PERC = 0.95
	const WAIT_CUTOFF = 3 * 1000000 // 3 seconds

	obj := c.dbusConn.Object("org.mpris.MediaPlayer2.spotify", "/org/mpris/MediaPlayer2")

	done := ctx.Done()

	started := false

	for {
		//Check for cancelation
		select {
		case <-done:
			log.Println("Waiter canceled")
			return
		default:
		}

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
		if pos == 0 && started {
			log.Println("FINISHED POS==0")
			c.onPlaybackFinished()
			return
		} else if rem < WAIT_CUTOFF && rem > 300000 {
			//Check every 0.1 seconds
			started = true
			log.Println("Waiting for 100ms")
			time.Sleep(100 * time.Millisecond)
		} else if rem < 30 {
			log.Println("FINISHED rem < 30")
			c.onPlaybackFinished()
			return
		} else {
			started = true
			log.Printf("Waiting for %v\n", time.Duration(float64(rem)*WAIT_PERC)*time.Microsecond)
			time.Sleep(time.Duration(float64(rem)*WAIT_PERC) * time.Microsecond)
		}

	}
}
