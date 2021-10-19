package contextaware

import (
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReader(t *testing.T) {

}

func TestReader(t *testing.T) {
	net.Dial

	cr := NewReader(pr)
	assert.IsType(t, readerWithSetDeadline{}, cr)

	cw := NewWriter(pw)
	assert.IsType(t, writerWithSetDeadline{}, cw)
}
