package fakes

import (
	"context"
	"sync"
)

type FakeMailSender struct {
	mu      sync.Mutex
	mailBox map[string][]string // email -> tokens
}

func NewFakeMailSender() *FakeMailSender {
	return &FakeMailSender{mailBox: make(map[string][]string)}
}

func (f *FakeMailSender) SendVerificationEmail(_ context.Context, toEmail, token string) error {
	f.mu.Lock()

	f.mailBox[toEmail] = append(f.mailBox[toEmail], token)

	f.mu.Unlock()
	return nil
}

func (f *FakeMailSender) LastToken(toEmail string) (string, bool) {
	f.mu.Lock()

	tokens, ok := f.mailBox[toEmail]
	if !ok || len(tokens) == 0 {
		return "", false
	}

	f.mu.Unlock()
	return tokens[len(tokens)-1], true
}

func (f *FakeMailSender) Reset() {
	f.mu.Lock()
	f.mailBox = make(map[string][]string)
	f.mu.Unlock()
}
