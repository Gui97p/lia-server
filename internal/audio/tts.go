package audio

import (
	"bytes"
	"context"

	edgetts "github.com/foresturquhart/edge-tts"
)

type TTSClient interface {
	Synthesize(ctx context.Context, text string) ([]byte, error)
}

type EdgeTTSClient struct {
	Voice string
}

func NewEdgeTTSClient(voice string) *EdgeTTSClient {
	return &EdgeTTSClient{Voice: voice}
}

func (c *EdgeTTSClient) Synthesize(ctx context.Context, text string) ([]byte, error) {
	config := edgetts.DefaultConfig()
	config.Voice = c.Voice

	comm, err := edgetts.NewCommunicate(text, config)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = comm.Stream(ctx, func(chunk edgetts.TTSChunk) error {
		if chunk.Type == edgetts.ChunkTypeAudio {
			buf.Write(chunk.Data)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
