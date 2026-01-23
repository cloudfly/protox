package protox

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBSON(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	timestamp := Timestamp{
		Millis: uint64(time.Now().UnixMilli()),
	}

	data, err := timestamp.MarshalBSON()
	require.NoError(err)

	var timestamp2 Timestamp
	require.NoError(timestamp2.UnmarshalBSON(data))
	assert.Equal(timestamp, timestamp2)
}
