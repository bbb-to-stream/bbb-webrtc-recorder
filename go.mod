module github.com/bigbluebutton/bbb-webrtc-recorder

go 1.21.0

toolchain go1.24.1

replace github.com/at-wat/ebml-go => ./pkg/ebml-go

require (
	github.com/AlekSi/pointer v1.2.0
	github.com/at-wat/ebml-go v0.16.0
	github.com/crazy-max/gonfig v0.6.0
	github.com/gomodule/redigo v1.8.9
	github.com/google/uuid v1.6.0
	github.com/kr/pretty v0.3.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pion/interceptor v0.1.25
	github.com/pion/rtcp v1.2.15
	github.com/pion/rtp v1.8.11
	github.com/pion/sdp/v3 v3.0.10
	github.com/pion/webrtc/v3 v3.2.24
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.15.0
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/pflag v1.0.5
	github.com/titanous/json5 v1.0.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/BurntSushi/toml v1.1.0 // indirect
	github.com/aler9/gortsplib v1.0.1 // indirect
	github.com/aler9/gortsplib/v2 v2.2.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bluenviron/gortsplib/v4 v4.12.3 // indirect
	github.com/bluenviron/mediacommon v1.14.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/pion/datachannel v1.5.5 // indirect
	github.com/pion/dtls/v2 v2.2.7 // indirect
	github.com/pion/ice/v2 v2.3.11 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/mdns v0.0.8 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/sctp v1.8.8 // indirect
	github.com/pion/srtp/v2 v2.0.18 // indirect
	github.com/pion/stun v0.6.1 // indirect
	github.com/pion/transport/v2 v2.2.3 // indirect
	github.com/pion/turn/v2 v2.1.3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
