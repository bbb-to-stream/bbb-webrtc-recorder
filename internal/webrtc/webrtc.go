package webrtc

import (
	"context"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/bigbluebutton/bbb-webrtc-recorder/internal/config"
	"github.com/bigbluebutton/bbb-webrtc-recorder/internal/webrtc/recorder"
	"github.com/bigbluebutton/bbb-webrtc-recorder/internal/webrtc/utils"
	"github.com/pion/interceptor"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v3"
	log "github.com/sirupsen/logrus"

	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/description"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
)

type WebRTC struct {
	ctx context.Context
	cfg config.WebRTC
	pc  *webrtc.PeerConnection
}

func NewWebRTC(ctx context.Context, cfg config.WebRTC) *WebRTC {
	return &WebRTC{ctx: ctx, cfg: cfg}
}

// func (w *WebRTC) streamToRTSP(username string) {
// 	log.Info("Stream to", username)

// 	//Potong username jadi hanya nim aja
// 	//4201250006. Jeremia Manurung
// 	s1 := strings.Split(username, ". ")

// 	forma := &format.VP8{
// 		PayloadTyp: 96,
// 	}
// 	desc := &description.Session{
// 		Medias: []*description.Media{{
// 			Type:    description.MediaTypeVideo,
// 			Formats: []format.Format{forma},
// 		}},
// 	}

// 	client := gortsplib.Client{}
// 	err := client.StartRecording("rtsp://localhost:8554/stream/"+s1[0], desc)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer client.Close()

// 	for {
// 		packet, _, err := track.ReadRTP()
// 		if err != nil {
// 			log.Error("Error reading RTP packet:", err)
// 			break
// 		}

// 		err = client.WritePacketRTP(desc.Medias[0], packet)
// 		if err != nil {
// 			log.Error("Error sending RTP packet to RTSP:", err)
// 			break
// 		}
// 	}
// }

func (w WebRTC) Init(
	offer webrtc.SessionDescription,
	r recorder.Recorder,
	username string,
	connStateCallbackFn func(state webrtc.ICEConnectionState),
	flowCallbackFn func(isFlowing bool, videoTimestamp time.Duration, closed bool),
) webrtc.SessionDescription {
	// Prepare the configuration
	cfg := webrtc.Configuration{
		ICEServers:   w.cfg.ICEServers,
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
	}

	// Create a MediaEngine object to configure the supported codec
	m := &webrtc.MediaEngine{}

	sdpOffer := sdp.SessionDescription{}

	if err := sdpOffer.Unmarshal([]byte(offer.SDP)); err != nil {
		panic(err)
	}

	// Setup the codecs you want to use.
	// Only support VP8 and OPUS, this makes our WebM muxer code simpler
	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8, ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
		PayloadType:        96,
	}, webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}
	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus, ClockRate: 48000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
		PayloadType:        111,
	}, webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	}

	se := &webrtc.SettingEngine{}
	se.SetSRTPReplayProtectionWindow(1024)
	if err := se.SetEphemeralUDPPortRange(w.cfg.RTCMinPort, w.cfg.RTCMaxPort); err != nil {
		panic(err)
	}

	i := &interceptor.Registry{}
	// Use the default set of Interceptors
	if err := webrtc.RegisterDefaultInterceptors(m, i); err != nil {
		panic(err)
	}

	// Create the API object with the MediaEngine
	api := webrtc.NewAPI(
		webrtc.WithMediaEngine(m),
		webrtc.WithSettingEngine(*se),
		webrtc.WithInterceptorRegistry(i),
	)

	// Create a new RTCPeerConnection
	peerConnection, err := api.NewPeerConnection(cfg)
	if err != nil {
		panic(err)
	}
	w.pc = peerConnection

	for _, md := range sdpOffer.MediaDescriptions {
		switch md.MediaName.Media {
		case "audio":
			if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
				panic(err)
			}
		case "video":
			if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
				panic(err)
			}
		}
	}

	var streamPackets = make(chan *rtp.Packet)

	forma := &format.VP8{
		PayloadTyp: 96,
	}
	desc := &description.Session{
		Medias: []*description.Media{{
			Type:    description.MediaTypeVideo,
			Formats: []format.Format{forma},
		}},
	}

	client := gortsplib.Client{}

	s1 := strings.Split(username, ".+")
	err = client.StartRecording("rtsp://localhost:8554/stream/"+s1[0], desc)
	if err != nil {
		panic(err)
	}

	var addToStream = func(packet *rtp.Packet) {
		streamPackets <- packet
	}

	go func() {
		for {
			i := <-streamPackets
			// log.Info("Writing to stream...")
			err := client.WritePacketRTP(desc.Medias[0], i)
			if err != nil {
				log.Error("Error sending RTP packet to RTSP:", err)
				break
			}
		}
		client.Close()
	}()

	// Set a handler for when a new remote track starts, this handler copies inbound RTP packets,
	// replaces the SSRC and sends them back
	peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		isAudio := track.Kind() == webrtc.RTPCodecTypeAudio
		isVideo := track.Kind() == webrtc.RTPCodecTypeVideo

		if isAudio {
			r.SetHasAudio(true)
		}
		log.WithField("session", w.ctx.Value("session")).
			Infof("%s (%d) track started", track.Codec().RTPCodecCapability.MimeType, track.PayloadType())

		if isVideo {
			// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
			done := make(chan bool)
			defer func() {
				done <- true
			}()
			go func() {
				ticker := time.NewTicker(time.Second * 3)
				for {
					select {
					case <-done:
						ticker.Stop()
						return
					case <-ticker.C:
						errSend := peerConnection.WriteRTCP([]rtcp.Packet{
							&rtcp.PictureLossIndication{
								MediaSSRC: uint32(track.SSRC()),
							},
						})
						if errSend != nil {
							log.WithField("session", w.ctx.Value("session")).
								Error(errSend)
						}
					}
				}
			}()
		}

		rl, err := utils.NewReceiveLog(1024)
		if err != nil {
			panic(err)
		}

		if true {
			senderSSRC := rand.Uint32()
			done := make(chan bool)
			defer func() {
				done <- true
			}()
			go func() {
				ticker := time.NewTicker(time.Millisecond * 50)
				for {
					select {
					case <-done:
						ticker.Stop()
						return
					case <-ticker.C:
						missing := rl.MissingSeqNumbers(10)
						if missing == nil || len(missing) == 0 {
							continue
						}
						nack := &rtcp.TransportLayerNack{
							SenderSSRC: senderSSRC,
							MediaSSRC:  uint32(track.SSRC()),
							Nacks:      utils.NackPairs(missing),
						}
						errSend := peerConnection.WriteRTCP([]rtcp.Packet{nack})
						if errSend != nil {
							log.WithField("session", w.ctx.Value("session")).
								Error(errSend)
						}
					}
				}
			}()
		}

		jb := utils.NewJitterBuffer(w.cfg.JitterBuffer)
		var s1, s2 uint16

		if true {
			done := make(chan bool)
			defer func() {
				done <- true
			}()
			go func() {
				ticker := time.NewTicker(time.Millisecond * 100)
				var wasFlowing, isFlowing bool
				for {
					select {
					case <-done:
						ticker.Stop()
						if isVideo {
							flowCallbackFn(false, r.VideoTimestamp(), true)
						}
						return
					case <-ticker.C:
						if s1 == s2 {
							isFlowing = false
						} else {
							isFlowing = true
						}
						s1 = s2
						if isVideo {
							if isFlowing != wasFlowing {
								flowCallbackFn(isFlowing, r.VideoTimestamp(), false)
								if isFlowing {
									ticker.Reset(time.Millisecond * 1000)
								} else {
									ticker.Reset(time.Millisecond * 100)
								}
							}
						}
						wasFlowing = isFlowing
					}
				}
			}()
		}

		var seq int64
		su := utils.NewSequenceUnwrapper(16)
		for {
			// Read RTP packets being sent to Pion
			rtp, _, readErr := track.ReadRTP()
			if readErr != nil {
				if readErr == io.EOF {
					log.WithField("session", w.ctx.Value("session")).
						Infof("%s track stopped", track.Codec().RTPCodecCapability.MimeType)
					return
				}
				log.WithField("session", w.ctx.Value("session")).
					Error(readErr)
				return
			}
			seq = su.Unwrap(uint64(rtp.SequenceNumber))
			jb.Add(seq, rtp)
			rl.Add(rtp.SequenceNumber)
			packets := jb.NextPackets()

			if packets == nil {
				continue
			}

			for _, p := range packets {
				s2 = p.SequenceNumber
				switch {
				case isAudio:
					r.PushAudio(p)
				case isVideo:
					r.PushVideo(p)
					addToStream(p)
				}
			}
		}
	})

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.WithField("session", w.ctx.Value("session")).
			Infof("webrtc connection state changed: %s", connectionState.String())
		if connectionState > webrtc.ICEConnectionStateConnected {
			if err := peerConnection.Close(); err != nil {
				log.WithField("session", w.ctx.Value("session")).
					Error(err)
			}
		}
		connStateCallbackFn(connectionState)
	})

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		log.WithField("session", w.ctx.Value("session")).
			Error(err)
		return webrtc.SessionDescription{}
	}

	// Create an answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		log.WithField("session", w.ctx.Value("session")).
			Error(err)
		return webrtc.SessionDescription{}
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	// Output the answer in base64
	return *peerConnection.LocalDescription()
}

func (w WebRTC) Close() error {
	return w.pc.Close()
}
