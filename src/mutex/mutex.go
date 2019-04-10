package mutex

import (
	"sync"
)

var (
	mutexes  = make(map[int64]*sync.Mutex)
	mapMutex = new(sync.Mutex)
)

// Lock locks the mutex for a given proposal, after instanciating if it doesn't
// exist.
func Lock(number int64) {
	// Avoid multiple goroutines using the map at the same time.
	mapMutex.Lock()
	defer mapMutex.Unlock()

	// Lock the existing mutex if there's one.
	if m, exists := mutexes[number]; exists {
		m.Lock()
		return
	}

	// If there's no mutex for this proposal, create one then lock it.
	mutexes[number] = new(sync.Mutex)
	mutexes[number].Lock()
}

// Unlock unlocks the mutex for a given proposal.
// Panics if the mutex doesn't exist.
func Unlock(number int64) {
	// Avoid multiple goroutines using the map at the same time.
	mapMutex.Lock()
	defer mapMutex.Unlock()

	mutexes[number].Unlock()
}
