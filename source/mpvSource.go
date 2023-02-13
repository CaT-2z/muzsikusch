package source

import (
	"context"
	"fmt"
	"log"
	entry "muzsikusch/queue/entry"

	"github.com/dexterlb/mpvipc"
)

// Base for a source based on mvp.
// Implement the Play and Search methods to use as a source.
// DO NOT forget to call waitForEnd in the constructor.
// TODO: Maybe we can do this automatically?
type MpvSource struct {
	instance           mpvipc.Connection
	events             chan *mpvipc.Event
	stopChan           chan struct{}
	onPlaybackFinished func()
	isActive           bool
}

// Create a new MpvSource with the given path and context.
// The context is used to stop the event listener.
func NewMpvSource(path string, ctx context.Context) (*MpvSource, error) {
	instance := mpvipc.NewConnection(path)
	err := instance.Open()
	if err != nil {
		return nil, err
	}

	events, closer := instance.NewEventListener()

	src := &MpvSource{
		instance: *instance,
		events:   events,
		stopChan: closer,
	}

	//This CANNOT be called here, because of Go's "inheritance" system.
	//go src.waitForEnd(ctx)

	return src, nil
}

// Register a callback to be called when playback is finished.
func (s *MpvSource) Register(cb func()) {
	s.onPlaybackFinished = cb
}

// Stop the currently playing song.
func (s *MpvSource) Stop() error {
	s.isActive = false
	_, err := s.instance.Call("stop")
	return err
}

// Skip to the next song
func (s *MpvSource) Skip() error {
	s.isActive = false
	_, err := s.instance.Call("playlist-next", "force")
	return err
}

// Pause the currently playing song
func (s *MpvSource) Pause() error {
	return s.instance.Set("pause", true)
}

// Resume the currently paused song
func (s *MpvSource) Resume() error {
	return s.instance.Set("pause", false)
}

func (s *MpvSource) GetTimePos() (float32, error) {
	x, err := s.instance.Get("time-pos")
	return x.(float32), err
}

// Seek forward by the given amount
func (s *MpvSource) Forward(amm int) error {
	_, err := s.instance.Call("seek", amm)
	return err
}

// Seek backward by the given amount
func (s *MpvSource) Reverse(amm int) error {
	return s.Forward(-amm)
}

// Set the volume to the given value
func (s *MpvSource) SetVolume(vol int) error {
	return s.instance.Set("volume", vol)
}

// Get the current volume
func (s *MpvSource) GetVolume() (int, error) {
	vol, err := s.instance.Get("volume")
	if err != nil {
		return 0, err
	}
	return int(vol.(float64)), nil
}

// Toggle mute the player
func (s *MpvSource) Mute() error {
	_, err := s.instance.Call("cycle", "mute")
	return err
}

func (s *MpvSource) ResolveTitle(mid *entry.MusicID) (string, error) {
	return "", fmt.Errorf("cannot resolve title on mpv source")
}

// Wait for the playback to end.
// This method will not exit on the first event, but will call the callback
// and wait for the next event.
// This function will exit when the context is canceled.
func (s *MpvSource) waitForEnd(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-s.events:
			if event.Name == "end-file" && s.isActive {
				s.onPlaybackFinished()
				s.isActive = false
			}
		case <-s.stopChan:
			log.Println("MpvSource: Stopping event listener")
			return
		}
	}
}
