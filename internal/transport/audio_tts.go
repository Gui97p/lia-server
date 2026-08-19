package transport

import (
	"encoding/json"
	"net/http"
	"strings"
)

type AudioTTSPayload struct {
	Text string `json:"text"`
}

func (s *Server) handleAudioTTS(w http.ResponseWriter, r *http.Request) {
	_, ok := userFromContext(r.Context())
	if !ok {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	var payload AudioTTSPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.Text) == "" {
		http.Error(w, "text required", http.StatusBadRequest)
		return
	}

	audio, err := s.ttsClient.Synthesize(r.Context(), payload.Text)
	if err != nil {
		http.Error(w, "failed to synthesize speech", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "audio/mpeg")
	w.Write(audio)
}
