package httpclient

import (
	"go.uber.org/ratelimit"
	"log"
	"sync"
	"time"
)

type RateLimitSpec struct {
	Rate int32
	// The name or a text/template of the name that will be used to identify Limiter instance
	Domain TemplateWrapper
}

var limiters = make(map[string]ratelimit.Limiter)
var limitersMutex sync.Mutex

func (s *RateLimitSpec) Get(o interface{}) (ratelimit.Limiter, error) {
	domain, err := s.Domain.Apply(o)
	if err != nil {
		return nil, err
	}

	log.Print("Using rate limit domain:", domain)

	limitersMutex.Lock()
	defer limitersMutex.Unlock()

	if limiter, ok := limiters[domain]; ok {
		return limiter, nil
	} else {
		limiter := ratelimit.New(int(s.Rate))
		limiters[domain] = limiter
		return limiter, nil
	}
}

func limitRate(l ratelimit.Limiter) {
	if l == nil {
		return
	}

	before := time.Now()
	after := l.Take()

	delay := after.Sub(before)
	log.Print("Waited for", delay)
}
