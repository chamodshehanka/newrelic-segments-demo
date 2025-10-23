package handlers

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/newrelic/go-agent/v3/newrelic"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// ComputeUntraced handles /compute-untraced with no New Relic instrumentation.
// It simulates an internal delay (2-3s) and calls an external API (httpbin delay 2).
func ComputeUntraced(c *fiber.Ctx) error {
	requestID := c.Locals("requestid")
	// Log incoming headers for correlation/debugging
	log.Printf("%v: ComputeUntraced - start; headers: newrelic=%s traceparent=%s tracestate=%s X-Request-Id=%s",
		requestID, c.Get("newrelic"), c.Get("traceparent"), c.Get("tracestate"), c.Get("X-Request-Id"))

	// Simulate 2-3s internal processing
	r := time.Duration(2000+rand.Intn(1000)) * time.Millisecond
	time.Sleep(r)

	start := time.Now()
	// External call with no tracing
	resp, err := http.Get("https://httpbin.org/delay/2")
	if err != nil {
		log.Printf("%v: error calling external API: %v", requestID, err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "error", "error": err.Error()})
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.ReadAll(resp.Body)

	duration := time.Since(start)
	log.Printf("%v: ComputeUntraced - end (internal %s, external %s)", requestID, r.String(), duration.String())

	return c.JSON(fiber.Map{
		"status":            "OK",
		"internal_duration": r.Milliseconds(),
		"external_duration": duration.Milliseconds(),
	})
}

// ComputeTraced handles /compute-traced and demonstrates New Relic segments.
// It expects a New Relic transaction available in the request context (middleware
// from fibernewrelic will create it when distributed tracing headers are present).
// We create an internal segment for local processing and an external segment for the
// request to the external API. The New Relic agent will automatically honor
// distributed tracing headers sent by the caller.
func ComputeTraced(c *fiber.Ctx) error {
	requestID := c.Locals("requestid")
	// Log incoming headers so we can verify header propagation from chamod
	log.Printf("%v: ComputeTraced - start; headers: newrelic=%s traceparent=%s tracestate=%s X-Request-Id=%s",
		requestID, c.Get("newrelic"), c.Get("traceparent"), c.Get("tracestate"), c.Get("X-Request-Id"))

	// Get New Relic transaction from context
	txn := newrelic.FromContext(c.UserContext())
	if txn == nil {
		log.Printf("%v: no New Relic transaction found; this request will be recorded locally only", requestID)
		return ComputeUntraced(c)
	}

	// Start internal segment to measure local processing
	internalStart := time.Now()
	internalSeg := txn.StartSegment("NisansalaInternalProcessing")
	// Simulate 2-3s processing
	r := time.Duration(2000+rand.Intn(1000)) * time.Millisecond
	time.Sleep(r)
	internalSeg.End()
	internalDuration := time.Since(internalStart)

	// Prepare external request with the same context so the agent can inject tracing headers
	req, err := http.NewRequestWithContext(c.UserContext(), http.MethodGet, "https://httpbin.org/delay/2", nil)
	if err != nil {
		log.Printf("%v: error creating external request: %v", requestID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "error": err.Error()})
	}

	// Start an external segment. This will cause the New Relic agent to add distributed
	// tracing headers to the outbound request (so the called service could link traces).
	externalSeg := newrelic.StartExternalSegment(txn, req)
	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if externalSeg != nil {
		externalSeg.End()
	}
	if err != nil {
		log.Printf("%v: error performing external call: %v", requestID, err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "error", "error": err.Error()})
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.ReadAll(resp.Body)
	externalDuration := time.Since(start)

	// Add some attributes to the New Relic transaction to make durations visible
	txn.AddAttribute("nisansala.internal_ms", internalDuration.Milliseconds())
	txn.AddAttribute("nisansala.external_ms", externalDuration.Milliseconds())

	log.Printf("%v: ComputeTraced - end (internal %s, external %s)", requestID, r.String(), externalDuration.String())

	return c.JSON(fiber.Map{
		"status":            "OK",
		"internal_duration": r.Milliseconds(),
		"external_duration": externalDuration.Milliseconds(),
	})
}
