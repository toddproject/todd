package frame

import (
	"testing"

	. "gopkg.in/check.v1"
)

func TestFrame(t *testing.T) {
	TestingT(t)
}

type FrameSuite struct{}

var _ = Suite(&FrameSuite{})

func (s *FrameSuite) TestNew(c *C) {
	f := New("CCC")
	c.Check(f.Header.Len(), Equals, 0)
	c.Check(f.Command, Equals, "CCC")

	f = New("DDDD", "abc", "def")
	c.Check(f.Header.Len(), Equals, 1)
	k, v := f.Header.GetAt(0)
	c.Check(k, Equals, "abc")
	c.Check(v, Equals, "def")
	c.Check(f.Command, Equals, "DDDD")

	f = New("EEEEEEE", "abc", "def", "hij", "klm")
	c.Check(f.Command, Equals, "EEEEEEE")
	c.Check(f.Header.Len(), Equals, 2)
	k, v = f.Header.GetAt(0)
	c.Check(k, Equals, "abc")
	c.Check(v, Equals, "def")
	k, v = f.Header.GetAt(1)
	c.Check(k, Equals, "hij")
	c.Check(v, Equals, "klm")
}

func (s *FrameSuite) TestClone(c *C) {
	f1 := &Frame{
		Command: "AAAA",
	}

	f2 := f1.Clone()
	c.Check(f2.Command, Equals, f1.Command)
	c.Check(f2.Header, IsNil)
	c.Check(f2.Body, IsNil)

	f1.Header = NewHeader("aaa", "1", "bbb", "2", "ccc", "3")

	f2 = f1.Clone()
	c.Check(f2.Header.Len(), Equals, f1.Header.Len())
	for i := 0; i < f1.Header.Len(); i++ {
		k1, v1 := f1.Header.GetAt(i)
		k2, v2 := f2.Header.GetAt(i)
		c.Check(k1, Equals, k2)
		c.Check(v1, Equals, v2)
	}

	f1.Body = []byte{1, 2, 3, 4, 5, 6, 5, 4, 77, 88, 99, 0xaa, 0xbb, 0xcc, 0xff}
	f2 = f1.Clone()
	c.Check(len(f2.Body), Equals, len(f1.Body))
	for i := 0; i < len(f1.Body); i++ {
		c.Check(f1.Body[i], Equals, f2.Body[i])
	}
}
