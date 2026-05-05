package watcher

import (
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher monitors a config file for changes and triggers a reload callback.
type Watcher struct {
	path     string
	onChange func()
	watcher  *fsnotify.Watcher
	done     chan struct{}
}

// New creates a new Watcher for the given file path.
// onChange is called whenever the file is modified.
func New(path string, onChange func()) (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if err := fw.Add(path); err != nil {
		fw.Close()
		return nil, err
	}
	return &Watcher{
		path:     path,
		onChange: onChange,
		watcher:  fw,
		done:     make(chan struct{}),
	}, nil
}

// Start begins watching the file in a background goroutine.
func (w *Watcher) Start() {
	go func() {
		// debounce rapid successive events
		var debounce <-chan time.Time
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					debounce = time.After(100 * time.Millisecond)
				}
			case <-debounce:
				log.Printf("[watcher] config changed: %s", w.path)
				w.onChange()
			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				log.Printf("[watcher] error: %v", err)
			case <-w.done:
				return
			}
		}
	}()
}

// Stop shuts down the watcher.
func (w *Watcher) Stop() {
	close(w.done)
	w.watcher.Close()
}
