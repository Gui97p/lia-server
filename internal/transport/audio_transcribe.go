package transport

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/Gui97p/lia-server/internal/crypto"
)

type AudioTranscribeResponse struct {
	Text string `json:"text"`
}

func (s *Server) handleAudioTranscribe(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	if user.GroqAPIKeyEncrypted == nil {
		http.Error(w, "api key not set", http.StatusBadRequest)
		return
	}

	groqAPIKey, err := crypto.Decrypt(*user.GroqAPIKeyEncrypted, s.encryptionKey)
	if err != nil {
		http.Error(w, "invalid api key", http.StatusBadRequest)
		return
	}

	audio, err := io.ReadAll(io.LimitReader(r.Body, 25<<20))
	if err != nil {
		s.logger.Error("error on parsing body", "error", err)
		http.Error(w, "failed to parse body", http.StatusInternalServerError)
		return
	}
	if len(audio) == 0 {
		http.Error(w, "audio required", http.StatusBadRequest)
		return
	}

	format, ok := audioFormat(r)
	if !ok {
		http.Error(w, "format required", http.StatusBadRequest)
		return
	}

	text, err := s.transcriberClient.Transcribe(r.Context(), groqAPIKey, audio, format)
	if err != nil {
		s.logger.Error("error on transcribing audio", "error", err)
		http.Error(w, "failed to transcribe speech", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AudioTranscribeResponse{Text: text})
}

func audioFormat(r *http.Request) (string, bool) {
	if f := r.URL.Query().Get("format"); f != "" {
		return f, true
	}
	ct := r.Header.Get("Content-Type")
	if i := strings.Index(ct, ";"); i >= 0 {
		ct = strings.TrimSpace(ct[:i])
	}
	switch ct {
	case "audio/wav", "audio/x-wav":
		return "wav", true
	case "audio/mpeg", "audio/mp3":
		return "mp3", true
	case "audio/webm":
		return "webm", true
	case "audio/flac":
		return "flac", true
	case "audio/mp4", "audio/m4a":
		return "m4a", true
	default:
		return "", false
	}
}
