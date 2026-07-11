package queue

import (
	"database/sql"
	"log"
	"sync"

	"github.com/robertpelloni/vst_monster/crawler/parser"
)

// Pipeline manages the concurrent scraping and output processing
type Pipeline struct {
	Results       chan parser.StandardPlugin
	wg            sync.WaitGroup
	db            *sql.DB
	onPluginFound func(plugin parser.StandardPlugin) // Optional callback
}

func NewPipeline(db *sql.DB) *Pipeline {
	return &Pipeline{
		Results: make(chan parser.StandardPlugin, 100),
		db:      db,
	}
}

// SetCallback allows injecting a custom action per plugin
func (p *Pipeline) SetCallback(cb func(plugin parser.StandardPlugin)) {
	p.onPluginFound = cb
}

func (p *Pipeline) StartConsumers(numConsumers int) {
	for i := 0; i < numConsumers; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for plugin := range p.Results {
				if p.onPluginFound != nil {
					p.onPluginFound(plugin)
				} else {
					jsonStr, err := parser.ToJSON(plugin)
					if err == nil {
						log.Printf("Ingested: %s", jsonStr)
					}
				}
			}
		}()
	}
}

func (p *Pipeline) Close() {
	close(p.Results)
	p.wg.Wait()
}
