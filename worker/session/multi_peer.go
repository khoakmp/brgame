package session

import (
	"log"
	"net"
	"sync"
	"sync/atomic"

	"github.com/khoakmp/brgame/api"
	"github.com/khoakmp/brgame/worker/coordinator"
	"github.com/khoakmp/brgame/worker/game"
	"github.com/khoakmp/brgame/worker/relay"
	"github.com/khoakmp/brgame/worker/utils"
	"github.com/khoakmp/brgame/worker/vm"
	"github.com/khoakmp/brgame/worker/webrtc"
	"github.com/pion/rtp"
)

type SessionMultiPeer struct {
	id            string
	appName       string
	workerID      string
	peers         map[string]*webrtc.Peer
	lock          sync.RWMutex
	coordConn     coordinator.Conn
	inMessageChan chan *api.Message

	trackVA *webrtc.TrackVA

	videoChannel chan *rtp.Packet
	audioChannel chan *rtp.Packet
	inputChannel chan webrtc.Packet

	videoListener *net.UDPConn
	audioListener *net.UDPConn
	sycnListener  *net.TCPListener

	relayer        *relay.Relayer
	containerID    string
	peersReadyChan chan struct{}
	exitChan       chan struct{}
	exitFlag       uint32
	exitLock       sync.RWMutex

	sessionHub *Hub
}

func getActualAppname(appName string) string {
	if appName == "bloody-roar-2" {
		return "br2"
	}
	return "br2"
}
func NewSessionMultiPeer(workerID, appName, sessionID string, peerIDs []string,
	hub *Hub, conn coordinator.Conn) (*SessionMultiPeer, error) {

	videoChannel := make(chan *rtp.Packet, 400)
	audioChannel := make(chan *rtp.Packet, 400)
	inputChannel := make(chan webrtc.Packet, 100)

	videoListener, err := net.ListenUDP("udp", &net.UDPAddr{Port: 0})
	if err != nil {
		log.Printf("[%s] failed to create video udp listener,%s\n", sessionID, err)
		return nil, err
	}

	videoPort, err := utils.GetPort(videoListener.LocalAddr().String())
	if err != nil {
		return nil, err
	}
	audioListener, err := net.ListenUDP("udp", &net.UDPAddr{Port: 0})
	if err != nil {
		log.Printf("[%s] failed to create audio udp listener,%s\n", sessionID, err)
		return nil, err
	}
	audioPort, err := utils.GetPort(audioListener.LocalAddr().String())
	if err != nil {
		return nil, err
	}

	syncListener, err := net.ListenTCP("tcp", &net.TCPAddr{Port: 0})

	if err != nil {
		log.Printf("[%s] failed to create sync tcp listener,%s\n", sessionID, err)
		return nil, err
	}
	syncPort, err := utils.GetPort(syncListener.Addr().String())

	if err != nil {
		return nil, err
	}
	containerID := utils.RandString(10)
	app := getActualAppname(appName)
	if err := vm.StartVM(containerID, app, videoPort, audioPort, syncPort); err != nil {
		return nil, err
	}

	trackVA, err := webrtc.NewTrackVA(sessionID, videoChannel, audioChannel, "stream")
	if err != nil {
		return nil, err
	}

	peersReadyChan := make(chan struct{}, len(peerIDs))

	session := SessionMultiPeer{
		id:             sessionID,
		appName:        app,
		workerID:       workerID,
		peers:          make(map[string]*webrtc.Peer),
		inMessageChan:  make(chan *api.Message),
		videoChannel:   videoChannel,
		audioChannel:   audioChannel,
		inputChannel:   inputChannel,
		peersReadyChan: peersReadyChan,
		exitChan:       make(chan struct{}),
		coordConn:      conn,
		lock:           sync.RWMutex{},
		trackVA:        trackVA,
		videoListener:  videoListener,
		audioListener:  audioListener,
		sycnListener:   syncListener,
		containerID:    containerID,
		exitFlag:       0,
		sessionHub:     hub,
		relayer:        nil,
	}

	for i := 0; i < len(peerIDs); i++ {
		p, err := webrtc.NewPeer(sessionID, peerIDs[i], i,
			trackVA, inputChannel, peersReadyChan, session.handleSendMessage, session.RemovePeer)
		if err != nil {
			log.Printf("[%s] Failed to create RTCPeer, %s\n", sessionID, err)
			session.exit()
			return nil, err
		}

		//session.lock.Lock()
		session.peers[peerIDs[i]] = p
		//session.lock.Unlock()
	}
	log.Printf("[Session %s] Created With PeerNum: %d\n", sessionID, len(session.peers))

	log.Printf("[Session %s] Create Relayer [VideoPort: %d, AudioPort: %d, SyncInputPort: %d]\n", session.id,
		videoPort, audioPort, syncPort)

	session.relayer = relay.NewRelayer(sessionID, videoChannel, audioChannel, inputChannel, videoListener, audioListener, syncListener, game.TranslateKeyFnMap[app])

	go session.readLoop()
	go session.start()

	return &session, nil
}

func (s *SessionMultiPeer) readLoop() {
	for {
		select {
		case <-s.exitChan:
			return
		case msg := <-s.inMessageChan:
			switch msg.Type {
			case api.MessageSDP:
				s.lock.RLock()
				peer, ok := s.peers[msg.SenderID]
				s.lock.RUnlock()
				if !ok {
					continue
				}
				// TODO: handle error
				peer.SetRemoteDescription(msg.Payload)
			case api.MessageICECandidate:
				s.lock.RLock()
				peer, ok := s.peers[msg.SenderID]
				s.lock.RUnlock()
				if !ok {
					continue
				}
				// TODO: handle error
				peer.AddICECandidate(msg.Payload)

			}
		}
	}
}

func (s *SessionMultiPeer) start() {
	var cnt int = 0
LOOP:
	for {
		select {
		case <-s.exitChan:
			return
		case <-s.peersReadyChan:
			cnt++
			if cnt == len(s.peers) {
				break LOOP
			}
		}
	}
	log.Printf("[Session %s] Start Streaming from UDP listner\n", s.id)
	s.trackVA.StartStreaming()
}

func (s *SessionMultiPeer) sendSDP(peerID, sdp string) {
	//log.Printf("[Session %s] Send SDP of Worker to peer: %s\n", s.id, peerID)
	s.coordConn.WriteMessage(api.Message{
		SessionID:   s.id,
		ReceiverIDs: []string{peerID},
		SenderID:    s.workerID,
		Type:        api.MessageSDP,
		Payload:     sdp,
	})
}

func (s *SessionMultiPeer) sendICECandidate(clientID string, candidate string) {
	//log.Printf("[Session %s] Send ICE candidate of Worker to peer: %s\n", s.id, clientID)
	s.coordConn.WriteMessage(api.Message{
		SessionID:   s.id,
		SenderID:    s.workerID,
		ReceiverIDs: []string{clientID},
		Type:        api.MessageICECandidate,
		Payload:     candidate,
	})
}

func (s *SessionMultiPeer) handleSendMessage(clientID, messageType, payload string) {
	switch messageType {
	case api.MessageSDP:
		s.sendSDP(clientID, payload)
	case api.MessageICECandidate:
		s.sendICECandidate(clientID, payload)
	}
}

func (s *SessionMultiPeer) Receive(msg api.Message) {
	if atomic.LoadUint32(&s.exitFlag) == 1 {
		return
	}
	s.exitLock.RLock()
	defer s.exitLock.RUnlock()

	s.inMessageChan <- &msg
}

func (s *SessionMultiPeer) RemovePeer(peerID string) {
	s.lock.Lock()
	delete(s.peers, peerID)
	if len(s.peers) > 0 {
		s.lock.Unlock()
		return
	}
	s.lock.Unlock()
	s.exit()
}

func (s *SessionMultiPeer) exit() {
	if !atomic.CompareAndSwapUint32(&s.exitFlag, 0, 1) {
		return
	}
	s.exitLock.Lock()
	defer s.exitLock.Unlock()

	log.Printf("[Session %s] closing\n", s.id)
	close(s.exitChan)

	if err := vm.StopVM(s.containerID, s.appName); err != nil {
		log.Printf("[%s] Failed to stop vm,%s\n", s.id, err)
	}

	s.videoListener.Close()
	s.audioListener.Close()
	s.sycnListener.Close()

	s.relayer.Close()

	close(s.videoChannel)
	close(s.audioChannel)

	s.lock.RLock()
	for _, p := range s.peers {
		p.Close()
	}
	s.lock.RUnlock()
	close(s.inputChannel)
	s.sessionHub.RemoveSession(s.id)
}

func (s *SessionMultiPeer) Close() {
	s.exit()
}

func (s *SessionMultiPeer) Stats() Stats {
	return Stats{}
}

func (s *SessionMultiPeer) ID() string {
	return s.id
}
