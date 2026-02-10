package protox

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestBSON(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	timestamp := Timestamp{
		Millis: uint64(time.Now().UnixMilli()),
	}

	data, err := bson.Marshal(timestamp)
	require.NoError(err)

	t.Log(data)

	var timestamp2 Timestamp
	require.NoError(bson.Unmarshal(data, &timestamp2))
	assert.Equal(timestamp, timestamp2)
}
