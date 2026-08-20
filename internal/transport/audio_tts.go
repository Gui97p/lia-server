package transport

import (
	"encoding/json"
	"net/http"
	"strings"
)

type AudioTTSPayload struct {
	Text string `json:"text"`
}

// @Summary      Transformar texto em fala
// @Description  Recebe um texto e devolve o áudio correspondente (MP3), sintetizado via Edge TTS
// @Tags         audio
// @Accept       json
// @Produce      audio/mpeg
// @Param        payload  body  AudioTTSPayload  true  "Texto a sintetizar"
// @Success      200  {file}    binary  "Áudio MP3"
// @Failure      400  {string}  string  "invalid payload / text required"
// @Failure      401  {string}  string  "missing token / invalid token"
// @Failure      500  {string}  string  "failed to synthesize speech"
// @Security     BearerAuth
// @Router       /audio/speak [post]
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
