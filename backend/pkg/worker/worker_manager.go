package worker

import (
	"fmt"
	"sync"
)

// Manager gère tous les workers de téléchargement actifs
type Manager struct {
	workers map[string]*DownloadWorker
	mu      sync.RWMutex
}

// NewManager crée un nouveau gestionnaire de workers
func NewManager() *Manager {
	return &Manager{
		workers: make(map[string]*DownloadWorker),
	}
}

// Add ajoute un worker au gestionnaire
func (m *Manager) Add(id string, worker *DownloadWorker) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.workers[id] = worker
}

// Get récupère un worker par son ID
func (m *Manager) Get(id string) (*DownloadWorker, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	worker, exists := m.workers[id]
	return worker, exists
}

// Remove supprime un worker du gestionnaire
func (m *Manager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if worker, exists := m.workers[id]; exists {
		worker.Close()
		delete(m.workers, id)
	}
}

// Pause met en pause un téléchargement
func (m *Manager) Pause(id string) error {
	worker, exists := m.Get(id)
	if !exists {
		return fmt.Errorf("download not found")
	}
	return worker.Pause()
}

// Resume reprend un téléchargement
func (m *Manager) Resume(id string) error {
	worker, exists := m.Get(id)
	if !exists {
		return fmt.Errorf("download not found")
	}
	return worker.Resume()
}

// Cancel annule un téléchargement
func (m *Manager) Cancel(id string) error {
	worker, exists := m.Get(id)
	if !exists {
		return fmt.Errorf("download not found")
	}
	return worker.Cancel()
}

// Count retourne le nombre de workers actifs
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.workers)
}

// CloseAll ferme tous les workers
func (m *Manager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, worker := range m.workers {
		worker.Close()
	}
	m.workers = make(map[string]*DownloadWorker)
}
