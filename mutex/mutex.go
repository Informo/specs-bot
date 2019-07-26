package mutex

import (
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	mutexes  = make(map[int64]*sync.Mutex)
	mapMutex = new(sync.Mutex)
)

// Lock locks the mutex for a given proposal, after instantiating if it doesn't
// exist.
func Lock(number int64) {
	// Avoid multiple goroutines using the map at the same time.
	logrus.Debug("Getting a lock on the global mutex")
	mapMutex.Lock()
	logrus.Debug("Got a lock on the global mutex")

	// Retrieve the mutex for this number from the map, or create one if none exist.
	m, exists := mutexes[number]
	if !exists {
		m = new(sync.Mutex)
		mutexes[number] = m
	}

	// Don't keep the global mutex locked if no more access to the map is needed.
	mapMutex.Unlock()
	logrus.Debug("Unlocked the global mutex")

	// Lock the mutex for this number.
	logrus.WithField("number", number).Debugf("Getting a lock on the mutex at address %p", m)
	m.Lock()
	logrus.WithField("number", number).Debugf("Got a lock on the mutex at address %p", m)
}

// Unlock unlocks the mutex for a given proposal.
// Does nothing if the mutex doesn't exist.
func Unlock(number int64) {
	// Avoid multiple goroutines using the map at the same time.
	logrus.Debug("Getting a lock on the global mutex")
	mapMutex.Lock()
	logrus.Debug("Got a lock on the global mutex")

	// Retrieve the mutex for this number from the map, if one exists.
	m, exists := mutexes[number]

	// Don't keep the global mutex locked if no more access to the map is needed.
	mapMutex.Unlock()
	logrus.Debug("Unlocked the global mutex")

	// Unlock the mutex for this number if one exists (otherwise we can safely ignore it).
	if exists {
		logrus.WithField("number", number).Debugf("Unlocking the mutex at address %p", m)
		m.Unlock()
		logrus.WithField("number", number).Debugf("Unlocked the mutex at address %p", m)
	}
}
