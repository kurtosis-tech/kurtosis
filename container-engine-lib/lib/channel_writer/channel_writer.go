package channel_writer

// Enables extracting output that is being written to an io.Writer via a channel
type ChannelWriter struct {
	channel chan string
}

func NewChannelWriter(channel chan string) *ChannelWriter {
	return &ChannelWriter{
		channel: channel,
	}
}

func (writer *ChannelWriter) Write(p []byte) (n int, err error) {
	writer.channel <- string(p)
	return len(p), nil
}
