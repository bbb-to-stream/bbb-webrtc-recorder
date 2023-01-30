package recorder

import (
	"fmt"
	"github.com/bigbluebutton/bbb-webrtc-recorder/internal/config"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"path/filepath"
	"time"
)

type FlowCallbackFn func(isFlowing bool, keyframeSequence int64, videoTimestamp time.Duration)

type Recorder interface {
	Push(rtp *rtp.Packet, track *webrtc.TrackRemote)
	Close()
}

func NewRecorder(cfg config.Recorder, file string, fn FlowCallbackFn) (Recorder, error) {
	ext := filepath.Ext(file)
	dir := path.Clean(cfg.Directory)

	if stat, err := os.Stat(dir); os.IsNotExist(err) {
		log.Debug(stat)
		return nil, fmt.Errorf("directory does not exist %s", cfg.Directory)
	}

	file = path.Clean(dir + string(os.PathSeparator) + file)
	if stat, err := os.Stat(file); !os.IsNotExist(err) {
		log.Debug(stat)
		return nil, fmt.Errorf("file already exists %s", file)
	}

	var r Recorder
	var err error
	switch ext {
	case ".webm":
		r = NewWebmRecorder(file, fn)
	default:
		err = fmt.Errorf("unsupported format '%s'", ext)
	}
	return r, err
}
