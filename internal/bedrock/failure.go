package bedrock

import "sync"

type FailureSink struct {
	mu        sync.Mutex
	err       error
	onFailure func()
}

func NewFailureSink(onFailure func()) *FailureSink {
	return &FailureSink{onFailure: onFailure}
}

func (s *FailureSink) Set(err error) {
	if err == nil {
		return
	}
	s.mu.Lock()
	first := s.err == nil
	if s.err == nil {
		s.err = err
	}
	s.mu.Unlock()
	if first && s.onFailure != nil {
		s.onFailure()
	}
}

func (s *FailureSink) Err() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}
