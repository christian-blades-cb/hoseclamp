package firehose

import (
	"bytes"
	"io"
)

// ContainerLineChannelWriter is a io.Writer which takes incoming lines and sends them, along with some metadata, on a defined channel.
type ContainerLineChannelWriter struct {
	OutputChannel chan<- *ContainerLine
	Image         string
	ContainerId   string
	buf           *bytes.Buffer
}

func (cw *ContainerLineChannelWriter) toChannel(p []byte) {
	cw.OutputChannel <- &ContainerLine{
		RawLine:     p,
		Image:       cw.Image,
		ContainerId: cw.ContainerId,
	}
}

// Write breaks a incoming byte slice into lines and sends the resulting ContainerLines down the OutputChannel
func (cw *ContainerLineChannelWriter) Write(p []byte) (n int, err error) {
	// combine the incoming byteslice with whatever was leftover in the buffer
	if n, err = cw.buf.Write(p); err != nil {
		return
	}

	// send each line through the channel
	var line []byte
	for err == nil {
		line, err = cw.buf.ReadBytes(byte('\n'))
		if err == nil {
			cw.toChannel(line)
		}
	}

	// put whatever is leftover into the buffer
	cw.buf.Reset()

	if err == io.EOF {
		cw.buf.Write(line)
		err = nil
	}

	return
}
