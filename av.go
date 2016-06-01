package av

type SampleFormat int

const (
	U8 = SampleFormat(iota+1)
	S16
	S32
	FLT
	DBL
	U8P
	S16P
	S32P
	FLTP
	DBLP
	U32
)

func (self SampleFormat) BytesPerSample() int {
	switch self {
	case U8,U8P:
		return 1
	case S16,S16P:
		return 2
	case FLT,FLTP,S32,S32P,U32:
		return 4
	case DBL,DBLP:
		return 8
	default:
		return 0
	}
}

func (self SampleFormat) String() string {
	switch self {
	case U8:
		return "U8"
	case S16:
		return "S16"
	case S32:
		return "S32"
	case FLT:
		return "FLT"
	case DBL:
		return "DBL"
	case U8P:
		return "U8P"
	case S16P:
		return "S16P"
	case FLTP:
		return "FLTP"
	case DBLP:
		return "DBLP"
	case U32:
		return "U32"
	default:
		return "?"
	}
}

func (self SampleFormat) IsPlanar() bool {
	switch self {
	case S16P,S32P,FLTP,DBLP:
		return true
	default:
		return false
	}
}

type ChannelLayout uint64

const (
	CH_FRONT_CENTER = ChannelLayout(1<<iota)
	CH_FRONT_LEFT
	CH_FRONT_RIGHT
	CH_BACK_CENTER
	CH_BACK_LEFT
	CH_BACK_RIGHT
	CH_SIDE_LEFT
	CH_SIDE_RIGHT
	CH_LOW_FREQ
	CH_NR

	CH_MONO = ChannelLayout(CH_FRONT_CENTER)
	CH_STEREO = ChannelLayout(CH_FRONT_LEFT|CH_FRONT_RIGHT)
	CH_2_1 = ChannelLayout(CH_STEREO|CH_BACK_CENTER)
	CH_2POINT1 = ChannelLayout(CH_STEREO|CH_LOW_FREQ)
	CH_SURROUND = ChannelLayout(CH_STEREO|CH_FRONT_CENTER)
	CH_3POINT1 = ChannelLayout(CH_SURROUND|CH_LOW_FREQ)
	// TODO: add all channel_layout in ffmpeg
)

func (self ChannelLayout) Count() (n int) {
	for self != 0 {
		n++
		self = (self-1)&self
	}
	return
}

const (
	H264 = iota+0x264
	AAC
	PCM_MULAW
)

type BasicAudioCodecData struct {
	CodecType int
	CodecSampleRate int
	CodecChannelLayout ChannelLayout
	CodecSampleFormat SampleFormat
}

func (self BasicAudioCodecData) Type() int {
	return self.CodecType
}

func (self BasicAudioCodecData) IsAudio() bool {
	return true
}

func (self BasicAudioCodecData) IsVideo() bool {
	return false
}

func (self BasicAudioCodecData) SampleRate() int {
	return self.CodecSampleRate
}

func (self BasicAudioCodecData) ChannelLayout() ChannelLayout {
	return self.CodecChannelLayout
}

func (self BasicAudioCodecData) SampleFormat() SampleFormat {
	return self.CodecSampleFormat
}

func NewPCMMulawCodecData() AudioCodecData {
	return BasicAudioCodecData{
		CodecType: PCM_MULAW,
		CodecSampleFormat: S16,
		CodecChannelLayout: CH_MONO,
		CodecSampleRate: 8000,
	}
}

type CodecData interface {
	IsVideo() bool
	IsAudio() bool
	Type() int
}

type VideoCodecData interface {
	CodecData
	Width() int
	Height() int
}

type AudioCodecData interface {
	CodecData
	SampleFormat() SampleFormat
	SampleRate() int
	ChannelLayout() ChannelLayout
}

type H264CodecData interface {
	VideoCodecData
	AVCDecoderConfRecordBytes() []byte
	SPS() []byte
	PPS() []byte
}

type AACCodecData interface {
	AudioCodecData
	MPEG4AudioConfigBytes() []byte
	MakeADTSHeader(samples int, payloadLength int) []byte
}

type Muxer interface {
	WriteHeader([]CodecData) error
	WritePacket(int, Packet) error
	WriteTrailer() error
}

type Demuxer interface {
	ReadPacket() (int, Packet, error)
	Duration() float64
	Streams() []CodecData
}

type Packet struct {
	IsKeyFrame      bool
	Data            []byte
	Duration        float64
	CompositionTime float64
}

type AudioFrame struct {
	SampleRate int
	SampleFormat SampleFormat
	ChannelLayout ChannelLayout
	SampleCount int
	Data [][]byte
}

func (self AudioFrame) HasSameFormat(other AudioFrame) bool {
	if self.SampleRate != other.SampleRate {
		return false
	}
	if self.ChannelLayout != other.ChannelLayout {
		return false
	}
	if self.SampleFormat != other.SampleFormat {
		return false
	}
	return true
}

func (self AudioFrame) Slice(start int, end int) (out AudioFrame) {
	out = self
	out.SampleCount = end-start
	size := self.SampleFormat.BytesPerSample()
	for i := range out.Data {
		out.Data[i] = out.Data[i][start*size:end*size]
	}
	return
}

func (self AudioFrame) Concat(in AudioFrame) (out AudioFrame) {
	out = self
	out.SampleCount += in.SampleCount
	for i := range out.Data {
		out.Data[i] = append(out.Data[i], in.Data[i]...)
	}
	return
}

type AudioEncoder interface {
	CodecData() AudioCodecData
	Encode(AudioFrame) ([]Packet, error)
	//Flush() ([]Packet, error)
}

type AudioDecoder interface {
	Decode(Packet) (AudioFrame, error)
	//Flush() (AudioFrame, error)
}

type AudioResampler interface {
	Resample(AudioFrame) (AudioFrame, error)
}
