package queue

import (
	"log"
	"sync"

	"github.com/robertpelloni/vst_monster/crawler/parser"
)

// Pipeline manages the concurrent scraping and output processing
type Pipeline struct {
	Results chan parser.StandardPlugin
	wg      sync.WaitGroup
}

func NewPipeline() *Pipeline {
	return &Pipeline{
		Results: make(chan parser.StandardPlugin, 100),
	}
}

func (p *Pipeline) StartConsumers() {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for plugin := range p.Results {
			jsonStr, err := parser.ToJSON(plugin)
			if err == nil {
				// Standard output for now (standalone per roadmap)
				log.Printf("Ingested: %s", jsonStr)
			}
		}
	}()
}

func (p *Pipeline) Close() {
	close(p.Results)
	p.wg.Wait()
}
