package audio

import (
	"bytes"
	"context"
	"regexp"
	"strings"

	edgetts "github.com/foresturquhart/edge-tts"
)

var emojiPattern = regexp.MustCompile(`[\x{1F300}-\x{1FAFF}\x{2600}-\x{27BF}\x{2190}-\x{21FF}\x{2B00}-\x{2BFF}\x{FE0F}]`)
var extraSpacePattern = regexp.MustCompile(`[ \t]{2,}`)

func stripEmoji(text string) string {
	stripped := emojiPattern.ReplaceAllString(text, "")
	return strings.TrimSpace(extraSpacePattern.ReplaceAllString(stripped, " "))
}

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

	comm, err := edgetts.NewCommunicate(stripEmoji(text), config)
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
