package main

import (
	"bytes"
	"sync"

	flvtag "github.com/yutopp/go-flv/tag"
)

type Pubsub struct {
	srv  *RelayService
	name string

	pub  *Pub
	subs []*Sub

	m sync.Mutex
}

func NewPubsub(srv *RelayService, name string) *Pubsub {
	return &Pubsub{
		srv:  srv,
		name: name,

		subs: make([]*Sub, 0),
	}
}

func (pb *Pubsub) Deregister() error {
	pb.m.Lock()
	defer pb.m.Unlock()

	for _, sub := range pb.subs {
		_ = sub.Close()
	}

	return pb.srv.RemovePubsub(pb.name)
}

func (pb *Pubsub) Pub() *Pub {
	pub := &Pub{
		pb: pb,
	}

	pb.pub = pub

	return pub
}

func (pb *Pubsub) Sub() *Sub {
	pb.m.Lock()
	defer pb.m.Unlock()

	sub := &Sub{}

	// TODO: Implement more efficient resource management
	pb.subs = append(pb.subs, sub)

	return sub
}

type Pub struct {
	pb *Pubsub

	avcSeqHeader *flvtag.FlvTag
	lastKeyFrame *flvtag.FlvTag
}

// Publish is responsible for publishing FLV data to subscribers
// It checks the type of incoming data (Audio, Script, Video)
// If it's video data, it checks whether it's a sequence header or a keyframe -> Used to init stream
// Sequence headers and keyframes are stored and sent to uninitialized subscribers
// to ensure correct video playback
func (p *Pub) Publish(flv *flvtag.FlvTag) error {
	// Switch on the type of data in the flv tag
	switch flv.Data.(type) {
	case *flvtag.AudioData, *flvtag.ScriptData:
		// If it's audio data or script data, simply clone and send to all subscribers
		for _, sub := range p.pb.subs {
			_ = sub.onEvent(cloneView(flv))
		}

	case *flvtag.VideoData:
		// If it's video data, check if it's a sequence header or a keyframe
		d := flv.Data.(*flvtag.VideoData)
		// If it's a sequence header, store it
		if d.AVCPacketType == flvtag.AVCPacketTypeSequenceHeader {
			p.avcSeqHeader = flv
		}

		// If it's a key frame, store it
		if d.FrameType == flvtag.FrameTypeKeyFrame {
			p.lastKeyFrame = flv
		}

		// Send the sequence header, keyframe, and current video data to all subscribers
		for _, sub := range p.pb.subs {
			// If the subscriber hasn't been initialized, send the header and keyframe first
			if !sub.initialized {
				if p.avcSeqHeader != nil {
					_ = sub.onEvent(cloneView(p.avcSeqHeader))
				}
				if p.lastKeyFrame != nil {
					_ = sub.onEvent(cloneView(p.lastKeyFrame))
				}
				sub.initialized = true
				continue
			}

			// Send the current video data to the subscriber
			_ = sub.onEvent(cloneView(flv))
		}

	default:
		// If it's an unexpected data type, panic
		panic("unexpected")
	}

	return nil
}

// Close deregisters the publisher from the room, stopping its publishing
func (p *Pub) Close() error {
	return p.pb.Deregister()
}

// Sub represents a Subscriber that receives data
type Sub struct {
	initialized bool
	closed      bool

	lastTimestamp uint32
	eventCallback func(*flvtag.FlvTag) error
}

// onEvent handles an event (an FLV tag) by sending it to the subscriber's callback
// If the subscriber is closed, it ignores the event
// If this is the first event the subscriber has received, it stores the timestamp
// It then adjusts the timestamp to be relative to the first timestamp
func (s *Sub) onEvent(flv *flvtag.FlvTag) error {
	// If the subscriber is closed, ignore the event
	if s.closed {
		return nil
	}

	// If this is the first event, store the timestamp
	if flv.Timestamp != 0 && s.lastTimestamp == 0 {
		s.lastTimestamp = flv.Timestamp
	}
	// Make the timestamp relative to the first timestamp
	flv.Timestamp -= s.lastTimestamp

	// Send the event to the subscriber's callback
	return s.eventCallback(flv)
}

// Close sets the subscriber as closed, stopping it from receiving any more events
func (s *Sub) Close() error {
	if s.closed {
		return nil
	}

	s.closed = true

	return nil
}

// cloneView creates a deep copy of an FlvTag
// This is needed because the FlvTag contains pointers, and modifying the copy inside a function
// would also affect the original outside the function
func cloneView(flv *flvtag.FlvTag) *flvtag.FlvTag {
	// Copy the FlvTag
	v := *flv

	// Deep copy the data inside the FlvTag
	switch flv.Data.(type) {
	case *flvtag.AudioData:
		// If it's AudioData, copy the AudioData
		dCloned := *v.Data.(*flvtag.AudioData)
		v.Data = &dCloned

		dCloned.Data = bytes.NewBuffer(dCloned.Data.(*bytes.Buffer).Bytes())

	case *flvtag.VideoData:
		// If it's VideoData, copy the VideoData
		dCloned := *v.Data.(*flvtag.VideoData)
		v.Data = &dCloned

		dCloned.Data = bytes.NewBuffer(dCloned.Data.(*bytes.Buffer).Bytes())

	case *flvtag.ScriptData:
		// If it's ScriptData, copy the ScriptData
		dCloned := *v.Data.(*flvtag.ScriptData)
		v.Data = &dCloned

	default:
		panic("unreachable")
	}

	return &v
}
