package frame

import (
	"bytes"
	"strings"

	. "gopkg.in/check.v1"
)

type WriterSuite struct{}

var _ = Suite(&WriterSuite{})

func (s *WriterSuite) TestWrites(c *C) {
	var frameTexts = []string{
		"CONNECT\nlogin:xxx\npasscode:yyy\n\n\x00",

		"SEND\n" +
			"destination:/queue/request\n" +
			"tx:1\n" +
			"content-length:5\n" +
			"\n\x00\x01\x02\x03\x04\x00",

		"SEND\ndestination:x\n\nABCD\x00",

		"SEND\ndestination:x\ndodgy\\nheader\\c:abc\\n\\c\n\n123456\x00",
	}

	for _, frameText := range frameTexts {
		writeToBufferAndCheck(c, frameText)
	}
}

func writeToBufferAndCheck(c *C, frameText string) {
	reader := NewReader(strings.NewReader(frameText))

	frame, err := reader.Read()
	c.Assert(err, IsNil)
	c.Assert(frame, NotNil)

	var b bytes.Buffer
	var writer = NewWriter(&b)
	err = writer.Write(frame)
	c.Assert(err, IsNil)
	newFrameText := b.String()
	c.Check(newFrameText, Equals, frameText)
	c.Check(b.String(), Equals, frameText)
}
