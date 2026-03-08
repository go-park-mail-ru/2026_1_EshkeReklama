package session

import "time"

// очищаем просроченные (пока по простому)
func (m *Manager) cleanupExpired() {
	now := time.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	for id, sess := range m.sessions {
		if now.After(sess.ExpiresAt) {
			delete(m.sessions, id)
		}
	}
}

func (m *Manager) StartCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			<-ticker.C
			m.cleanupExpired()
		}
	}()
}
