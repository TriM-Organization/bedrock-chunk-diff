package timeline

import (
	"context"
	"sync"

	"maps"

	"github.com/TriM-Organization/bedrock-chunk-diff/define"
)

// InProgressSession holds the timelines that are still in use.
type InProgressSession struct {
	mu      *sync.Mutex
	closed  bool
	session map[define.DimSubChunk]context.Context
}

// NewInProgressSession returns a new InProgressSession
func NewInProgressSession() *InProgressSession {
	return &InProgressSession{
		mu:      new(sync.Mutex),
		session: make(map[define.DimSubChunk]context.Context),
	}
}

// Require loads a new session which on pos, and ensure there is only
// one thread is using a timeline from the same sub chunk.
//
// If there is one thread is using the target timeline, then calling
// Require will blocking until they finish there using.
//
// Calling releaseFunc to show you release this timeline, and then other
// thread could start to using them.
// If you get Require returned false, then that means the underlying
// database is closed.
func (i *InProgressSession) Require(pos define.DimSubChunk) (releaseFunc func(), success bool) {
	var cancelFunc context.CancelFunc

	for {
		i.mu.Lock()
		if i.closed {
			i.mu.Unlock()
			return nil, false
		}
		ctx, ok := i.session[pos]
		i.mu.Unlock()

		if ok {
			<-ctx.Done()
		}

		i.mu.Lock()
		{
			if i.closed {
				i.mu.Unlock()
				return nil, false
			}
			if _, ok = i.session[pos]; ok {
				i.mu.Unlock()
				continue
			}
			ctx, cancelFunc = context.WithCancel(context.Background())
			i.session[pos] = ctx
		}
		i.mu.Unlock()

		break
	}

	release := func() {
		i.mu.Lock()
		{
			cancelFunc()
			delete(i.session, pos)
			newMapping := make(map[define.DimSubChunk]context.Context)
			maps.Copy(newMapping, i.session)
			i.session = newMapping
		}
		i.mu.Unlock()
	}

	return sync.OnceFunc(release), true
}
