package webrtc

import (
	"log"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

type TrackVA struct {
	traceID      string
	videoChannel <-chan *rtp.Packet
	audioChannel <-chan *rtp.Packet
	videoTrack   *webrtc.TrackLocalStaticRTP
	audioTrack   *webrtc.TrackLocalStaticRTP
}

func NewTrackVA(traceID string, videoChannel, audioChannel <-chan *rtp.Packet, streamID string,
) (*TrackVA, error) {
	videoTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{
		MimeType: webrtc.MimeTypeVP8,
	}, "video", streamID)
	if err != nil {
		log.Printf("[Session %s] Failed to create video track, %s\n", traceID, err.Error())
		return nil, err
	}

	audioTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{
		MimeType: webrtc.MimeTypeOpus,
	}, "audio", streamID)
	if err != nil {
		log.Printf("[%s] Failed to create audio track,%s\n", traceID, err.Error())
		return nil, err
	}
	log.Printf("[Session %s] Create Video-Audio Track OK\n", traceID)

	trackVA := &TrackVA{
		traceID:      traceID,
		videoChannel: videoChannel,
		audioChannel: audioChannel,
		videoTrack:   videoTrack,
		audioTrack:   audioTrack,
	}
	return trackVA, nil
}

func (t *TrackVA) StartStreaming() {
	go t.startStreamVideo()
	go t.startStreamAudio()
}
func (t *TrackVA) startStreamVideo() {
	for pkt := range t.videoChannel {
		if err := t.videoTrack.WriteRTP(pkt); err != nil {
			log.Printf("[%s] Failed write packet to VideoTrack %s\n", t.traceID, err.Error())
		}
	}

	log.Printf("[Session %s] Stop Writing video packet\n", t.traceID)
}

func (t *TrackVA) startStreamAudio() {
	for pkt := range t.audioChannel {
		if err := t.audioTrack.WriteRTP(pkt); err != nil {
			log.Printf("[%s] Failed write packet to AudioTrack %s\n", t.traceID, err.Error())
		}
	}
	log.Printf("[Session %s] Stop Writing audio packet\n", t.traceID)
}

func (t *TrackVA) TrackVideo() *webrtc.TrackLocalStaticRTP {
	return t.videoTrack
}

func (t *TrackVA) TrackAudio() *webrtc.TrackLocalStaticRTP {
	return t.audioTrack
}
