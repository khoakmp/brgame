package webrtc

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"sync"
	"sync/atomic"

	"github.com/khoakmp/brgame/api"
	"github.com/khoakmp/brgame/worker/utils"
	"github.com/pion/webrtc/v3"
)

/* var (
	webrtcSettings webrtc.SettingEngine
	settingOnce    sync.Once
) */

type Peer struct {
	sessionID     string
	clientID      string
	peerOrder     int
	conn          *webrtc.PeerConnection
	inputDataChan *webrtc.DataChannel
	closeChan     chan struct{}
	exitFlag      int32
}

type Packet struct {
	PeerOrder int
	Packet    *InputPacket
}

func NewPeer(sessionID, clientID string,
	peerOrder int,
	trackVA *TrackVA,
	inputChannel chan<- Packet,
	peersReadyChan chan<- struct{},
	handleSendMessage func(clientID, messageType, payload string), exitCb func(peerID string)) (*Peer, error) {
	/* m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		return nil, err
	}

	i := &interceptor.Registry{}
	if !settings.DisableDefaultInterceptors {
		if err := webrtc.RegisterDefaultInterceptors(m, i); err != nil {
			return nil, err
		}
	}

	settingOnce.Do(func() {
		settingEngine := webrtc.SettingEngine{}

		if settings.PortRange.Min > 0 && settings.PortRange.Max > 0 {
			if err := settingEngine.SetEphemeralUDPPortRange(settings.PortRange.Min, settings.PortRange.Max); err != nil {
				panic(err)
			}
		} else if settings.SinglePort > 0 {
			l, err := socket.NewSocketPortRoll("udp", settings.SinglePort)
			if err != nil {
				panic(err)
			}
			udpListener := l.(*net.UDPConn)
			log.Printf("[%s] Listening for WebRTC traffic at %s\n", sessionID, udpListener.LocalAddr())
			settingEngine.SetICEUDPMux(webrtc.NewICEUDPMux(nil, udpListener))
		}

		if settings.IceIpMap != "" {
			settingEngine.SetNAT1To1IPs([]string{settings.IceIpMap}, webrtc.ICECandidateTypeHost)
		}

		webrtcSettings = settingEngine
	})

	webRTCApi := webrtc.NewAPI(
		webrtc.WithMediaEngine(m),
		webrtc.WithInterceptorRegistry(i),
		webrtc.WithSettingEngine(webrtcSettings),
	) */

	conn, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	})

	if err != nil {
		log.Println("failed to create RTC Peer", err)
		return nil, err
	}

	videoSender, err := conn.AddTrack(trackVA.videoTrack)
	if err != nil {
		return nil, err
	}

	audioSender, err := conn.AddTrack(trackVA.audioTrack)
	if err != nil {
		return nil, err
	}

	inputCh, err := conn.CreateDataChannel("syncinput", nil)
	if err != nil {

		return nil, err
	}
	peer := &Peer{
		clientID:      clientID,
		sessionID:     sessionID,
		peerOrder:     peerOrder,
		conn:          conn,
		inputDataChan: inputCh,
		closeChan:     make(chan struct{}),
	}
	inputCh.OnMessage(func(msg webrtc.DataChannelMessage) {
		var inputPacket InputPacket
		err := json.Unmarshal(msg.Data, &inputPacket)

		if err != nil {
			log.Printf("[%s] Failed to parse input data, %s\n", sessionID, err)
			return
		}
		//log.Printf("[Session %s] Recv input from peer %s, payload:%s\n", peer.sessionID, peer.clientID, msg.Data)

		select {
		case <-peer.closeChan:
			return
		case inputChannel <- Packet{
			PeerOrder: peerOrder,
			Packet:    &inputPacket,
		}:
			//log.Println("send input packet to input channel")
			break
		}
		/* inputChannel <- Packet{
			PeerOrder: peerOrder,
			Packet:    &inputPacket,
		} */
	})

	conn.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}
		payload, err := utils.EncodeBase64(candidate.ToJSON())
		if err != nil {
			log.Printf("[%s] Failed to encode ice candidate,%s\n", clientID, err)
			return
		}
		//log.Printf("[Session %s, Peer %s] Receive ICE candidate from Stun \n", sessionID, clientID)
		handleSendMessage(clientID, api.MessageICECandidate, payload)
	})

	var onceReady sync.Once

	readyFunc := func() {
		peersReadyChan <- struct{}{}
	}

	conn.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		if state == webrtc.ICEConnectionStateConnected {
			log.Printf("[Session %s] Peer %s ICE connection change to state %d\n",
				peer.sessionID, peer.clientID, state)
			onceReady.Do(readyFunc)
			return
		}

		if state == webrtc.ICEConnectionStateFailed || state == webrtc.ICEConnectionStateDisconnected {
			log.Printf("[Session %s] Peer %s ICE connection change to state %d\n",
				peer.sessionID, peer.clientID, state)

			onceReady.Do(readyFunc)
			peer.Close()
			exitCb(clientID)
		}
	})

	go peer.readRTCPLoop(videoSender, "video")
	go peer.readRTCPLoop(audioSender, "audio")

	offer, err := peer.genLocalDescription()
	if err != nil {
		return nil, err
	}

	log.Printf("[Session %s] peer %s Gen local SDP OK\n", sessionID, clientID)

	handleSendMessage(clientID, api.MessageSDP, offer)
	return peer, nil
}

func (p *Peer) Close() {
	if !atomic.CompareAndSwapInt32(&p.exitFlag, 0, 1) {
		return
	}
	close(p.closeChan)
	if p.inputDataChan != nil {
		p.inputDataChan.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}
func (p *Peer) genLocalDescription() (string, error) {
	offer, err := p.conn.CreateOffer(nil)
	if err != nil {
		return "", err
	}

	if err := p.conn.SetLocalDescription(offer); err != nil {
		return "", err
	}

	return utils.EncodeBase64(offer)
}

func (p *Peer) readRTCPLoop(sender *webrtc.RTPSender, trackType string) {
	for {
		_, _, err := sender.ReadRTCP()
		if err != nil {
			log.Printf("[%s] Failed to read RTCP from %s sender,%s\n", p.sessionID, trackType, err)
		}
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) {
			log.Printf("[%s] Peer %s Stop Read RTCP for track %s\n", p.sessionID, p.clientID, trackType)
			return
		}
		// TODO: handle read packets
	}
}

func (p *Peer) SetRemoteDescription(sdp string) error {
	var sd webrtc.SessionDescription

	utils.DecodeBase64(sdp, &sd)

	return p.conn.SetRemoteDescription(sd)
}

func (p *Peer) AddICECandidate(ice string) error {
	var candiate webrtc.ICECandidateInit
	if err := utils.DecodeBase64(ice, &candiate); err != nil {
		return err
	}
	return p.conn.AddICECandidate(candiate)
}
