package sendamatic

// SendResponse repräsentiert die Antwort auf einen Send-Request
type SendResponse struct {
	StatusCode int
	Recipients map[string][2]interface{} // Email -> [StatusCode, MessageID]
}

// IsSuccess prüft, ob die gesamte Sendung erfolgreich war
func (r *SendResponse) IsSuccess() bool {
	return r.StatusCode == 200
}

// GetMessageID gibt die Message-ID für einen Empfänger zurück
func (r *SendResponse) GetMessageID(email string) (string, bool) {
	if info, ok := r.Recipients[email]; ok && len(info) >= 2 {
		if msgID, ok := info[1].(string); ok {
			return msgID, true
		}
	}
	return "", false
}

// GetStatus gibt den Status-Code für einen Empfänger zurück
func (r *SendResponse) GetStatus(email string) (int, bool) {
	if info, ok := r.Recipients[email]; ok && len(info) >= 1 {
		// Die API gibt float64 zurück, da JSON numbers als float64 dekodiert werden
		if status, ok := info[0].(float64); ok {
			return int(status), true
		}
	}
	return 0, false
}
