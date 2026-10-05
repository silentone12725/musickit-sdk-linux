// Package tracer is a no-op timing hook. The engine calls it around licence
// acquisition and stream start; replace the implementation to export latency
// traces without changing callers.
package tracer

import "context"

type Tracer struct{}

func FromContext(_ context.Context) *Tracer { return &Tracer{} }

func (*Tracer) RecordWebplaybackStart()  {}
func (*Tracer) RecordWebplaybackEnd()    {}
func (*Tracer) RecordCatalogFetchStart() {}
func (*Tracer) RecordCatalogFetchEnd()   {}
func (*Tracer) RecordLicenseStart()      {}
func (*Tracer) RecordLicenseEnd()        {}
func (*Tracer) RecordPlaybackReady()     {}
func (*Tracer) RecordRetry()             {}
func (*Tracer) RecordCBCSDialStart()     {}
func (*Tracer) RecordCBCSDialConnected() {}
func (*Tracer) RecordCBCSDownloadStart() {}
