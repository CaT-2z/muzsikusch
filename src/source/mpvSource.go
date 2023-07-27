package source

import (
	"context"
	"encoding/gob"
	"fmt"
	"github.com/blang/mpv"
	"log"
	"muzsikusch/src/queue/entry"
	"net/rpc"
	"os"
)

// Base for a source based on mvp.
// Implement the Play and Search methods to use as a source.
// DO NOT forget to call waitForEnd in the constructor.
// TODO: Maybe we can do this automatically?
// Let me rage here a little bit:
// - So we have an mpv library with a high level abstraction for connections
// - AND YET WE USE THE LOW LEVEL VARIANT SO NOW I NEED TO REWRITE EVERYTHING??
type MpvSource struct {
	client mpv.Client
	//instance           mpvipc.Connection
	//events             chan *mpvipc.Event
	//stopChan           chan struct{}
	onPlaybackFinished func()
	isActive           bool
}

var static_client *mpv.Client

// Create a new MpvSource with the given path and context.
// The context is used to stop the event listener.
func NewMpvSource(path string, ctx context.Context) (*MpvSource, error) {
	//instance := mpvipc.NewConnection(path)
	//err := instance.Open()
	//if err != nil {
	//	return nil, err
	//}
	//
	//events, closer := instance.NewEventListener()
	//
	if static_client == nil {

		log.Printf("Connecting to mpv rpc server " + os.Getenv("RPC_HOST"))

		gob.Register(map[string]interface{}{})

		connection, err := rpc.DialHTTP("tcp", os.Getenv("RPC_HOST"))
		if err != nil {
			return nil, err
		}
		if connection == nil {
			log.Printf("Couldnt connect to rpc")
		}

		static_client = mpv.NewClient(mpv.NewRPCClient(connection))
	}

	src := &MpvSource{
		client: *static_client,
		//instance: *instance,
		//events:   events,
		//stopChan: closer,
	}

	//This CANNOT be called here, because of Go's "inheritance" system.
	//go src.waitForEnd(ctx)

	return src, nil
}

// Stop the currently playing song.
func (s *MpvSource) Stop() error {
	s.isActive = false
	_, err := s.client.Exec("stop")
	return err
}

// Skip to the next song
func (s *MpvSource) Skip() error {
	s.isActive = false
	_, err := s.client.Exec("playlist-next", "force")
	return err
}

// Pause the currently playing song
func (s *MpvSource) Pause() error {
	err := s.client.SetPause(true)
	return err
}

// Resume the currently paused song
func (s *MpvSource) Resume() error {
	return s.client.SetPause(false)
}

func (s *MpvSource) Register(func()) {

}

func (s *MpvSource) GetTimePos() (float32, error) {
	x, err := s.client.GetFloatProperty("time-pos")
	if err != nil {
		return 0, err
	}
	return float32(x), err
}

// Seek forward by the given amount
func (s *MpvSource) Forward(amm int) error {
	err := s.client.Seek(amm, mpv.SeekModeRelative)
	return err
}

// Seek backward by the given amount
func (s *MpvSource) Reverse(amm int) error {
	return s.Forward(-amm)
}

// Set the volume to the given value
func (s *MpvSource) SetVolume(vol int) error {
	return s.client.SetProperty("volume", vol)
}

// Get the current volume
func (s *MpvSource) GetVolume() (int, error) {
	vol, err := s.client.GetFloatProperty("volume")
	if err != nil {
		return 0, err
	}
	return int(vol), nil
}

// Toggle mute the player
func (s *MpvSource) Mute() error {
	_, err := s.client.Exec("cycle", "mute")
	return err
}

func (s *MpvSource) ResolveTitle(mid *entry.MusicID) (string, error) {
	return "", fmt.Errorf("cannot resolve title on mpv source")
}

func (s *MpvSource) PlayUrl(url string) error {
	err := s.client.Loadfile(url, mpv.LoadFileModeReplace)
	// There is no war in Ba Sing Se
	if err == nil {
		s.isActive = true
	}
	return err
}

// Wait for the playback to end.
// This method will not exit on the first event, but will call the callback
// and wait for the next event.
// This function will exit when the context is canceled.
func (s *MpvSource) waitForEnd(ctx context.Context) {
}

//	for {
//		select {
//		case <-ctx.Done():
//			return
//		case event := <-s.events:
//			if event.Name == "end-file" && s.isActive {
//				s.onPlaybackFinished()
//				s.isActive = false
//			}
//		case <-s.stopChan:
//			log.Println("MpvSource: Stopping event listener")
//			return
//		}
//	}
//}
