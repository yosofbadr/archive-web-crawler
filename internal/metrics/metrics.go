package metrics

import (
	"sync/atomic"
	"time"
)

type Registry interface {
	IncPagesCrawled()
	IncLinksDiscovered(n int)
	ObserveRequest(duration time.Duration)
	PagesCrawled() uint64
	LinksDiscovered() uint64
}

type registry struct {
	pagesCrawled    atomic.Uint64
	linksDiscovered atomic.Uint64
}

func NewRegistry() Registry { return &registry{} }

func (r *registry) IncPagesCrawled()                { r.pagesCrawled.Add(1) }
func (r *registry) IncLinksDiscovered(n int)        { r.linksDiscovered.Add(uint64(n)) }
func (r *registry) ObserveRequest(_ time.Duration)  {}
func (r *registry) PagesCrawled() uint64            { return r.pagesCrawled.Load() }
func (r *registry) LinksDiscovered() uint64         { return r.linksDiscovered.Load() }
