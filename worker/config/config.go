package config

type Configuration struct {
	VideoBufferSize  int `yaml:"video_buffer_size,default:400"`
	AudioBufferSize  int `yaml:"audio_buffer_size,default:100"`
	InputChannelSize int `yaml:"input_channel_size,default:50"`
	VideoClockRate   int `yaml:"video_clock_rate,default:90"`
}

var AppConfig Configuration
