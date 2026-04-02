package detect

import (
	"context"
	"sync"
	"time"
)

type Summary struct {
	System    SystemInfo
	Providers []ProviderInfo
}

func Run(ctx context.Context) Summary {
	rctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var summary Summary
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		summary.System = DetectSystem(rctx)
	}()
	go func() {
		defer wg.Done()
		summary.Providers = DetectProviders(rctx)
	}()
	wg.Wait()

	for i := range summary.Providers {
		provider := &summary.Providers[i]
		if !provider.Found || !provider.NeedsModel || rctx.Err() != nil {
			continue
		}

		switch provider.Name {
		case "ollama":
			provider.Models = discoverOllamaModels(rctx)
		case "lmstudio":
			provider.Models = discoverLMStudioModels(rctx)
		}
	}

	return summary
}
