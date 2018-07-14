package manualNotify

type channelWriter struct {
	channel chan byte
}

func NewChannelWriter() *channelWriter {
	return &channelWriter{
		channel: make(chan byte, 512),
	}
}

func (writer *channelWriter) Get() chan byte {
	return writer.channel
}

func (writer *channelWriter) Write(data []byte) (int, error) {
	ctr := 0
	for _, single_byte := range data {
		writer.channel <- single_byte
		ctr++
	}
	return ctr, nil
}

func (writer *channelWriter) Close() error {
	close(writer.channel)
	return nil
}
