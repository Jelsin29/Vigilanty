package checker

import "sync"

var initOnce sync.Once

func Init() {
	initOnce.Do(func() {
		Register("shell", newShellChecker)
		Register("command", newShellChecker)
		Register("ai-review", newAIReviewChecker)
		Register("ai", newAIReviewChecker)
		Register("prompt", newAIReviewChecker)
	})
}
