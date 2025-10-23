package handlers

import (
	"io"
	"math/rand"
	"net/http"
	"time"

	"chamod/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/newrelic/go-agent/v3/newrelic"
)

func UntracedHandler(c *fiber.Ctx) error {
	// Safely extract request id
	requestID := ""
	if rid, ok := c.Locals("requestid").(string); ok {
		requestID = rid
	}
	utils.Logger.Info(requestID, "UntracedHandler - start")

	// Simulate 100-300ms internal processing
	r := time.Duration(100+rand.Intn(200)) * time.Millisecond
	time.Sleep(r)

	start := time.Now()
	// Build request so we can forward X-Request-Id for easier correlation
	url := "http://localhost:8081/compute-untraced"
	req, err := http.NewRequestWithContext(c.UserContext(), http.MethodGet, url, nil)
	if err != nil {
		utils.Logger.Error(requestID, "error creating request: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "error": err.Error()})
	}

	// Forward X-Request-Id if present
	if requestID != "" {
		req.Header.Set("X-Request-Id", requestID)
	}
	// Log outbound for verification (will show if NR headers were injected)
	utils.Logger.Info(requestID, "Untraced outbound request url=%s headers=%v", req.URL.String(), req.Header)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		utils.Logger.Error(requestID, "error calling nisansala untraced: %v", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "error", "error": err.Error()})
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.ReadAll(resp.Body)

	duration := time.Since(start)
	utils.Logger.Info(requestID, "UntracedHandler - end (internal %s, downstream %s)", r.String(), duration.String())

	return c.JSON(fiber.Map{
		"status":              "OK",
		"internal_duration":   r.Milliseconds(),
		"downstream_duration": duration.Milliseconds(),
	})
}

func TracedHandler(c *fiber.Ctx) error {
	// Safely extract request id
	requestID := ""
	if rid, ok := c.Locals("requestid").(string); ok {
		requestID = rid
	}
	utils.Logger.Info(requestID, "TracedHandler - start")

	// Get New Relic transaction from the fiber context.
	txn := newrelic.FromContext(c.UserContext())
	if txn == nil {
		utils.Logger.Warn(requestID, "no new relic transaction found, falling back to untraced handler")
		return UntracedHandler(c)
	} else {
		traceMetadata := txn.GetTraceMetadata()
		traceID := traceMetadata.TraceID
		spanID := traceMetadata.SpanID
		sampled := txn.IsSampled()
		utils.Logger.Debug(requestID, "Request TraceID: %s, SpanID: %s, Sampled: %v", traceID, spanID, sampled)
	}

	// Start an internal segment for local processing and measure duration manually
	internalStart := time.Now()
	internalSeg := txn.StartSegment("ChamodInternalProcessing")
	// Simulate 100-300ms internal processing
	r := time.Duration(100+rand.Intn(200)) * time.Millisecond
	time.Sleep(r)
	defer internalSeg.End()
	internalDuration := time.Since(internalStart)

	// Prepare an outbound request to nisansala traced endpoint. Using the transaction
	// with StartExternalSegment will cause the agent to add distributed tracing headers
	// to the outbound HTTP request automatically.
	url := "http://localhost:8081/compute-traced"
	req, err := http.NewRequestWithContext(c.UserContext(), http.MethodGet, url, nil)
	if err != nil {
		utils.Logger.Error(requestID, "error creating request: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "error": err.Error()})
	}

	// Forward X-Request-Id for easier correlation in logs
	if requestID != "" {
		req.Header.Set("X-Request-Id", requestID)
	}

	// Start external segment which wraps the HTTP call and injects distributed tracing headers
	externalSeg := newrelic.StartExternalSegment(txn, req)
	// Log outbound url and headers so you can confirm the agent injected DT headers
	utils.Logger.Info(requestID, "Traced outbound request url=%s headers=%v", req.URL.String(), req.Header)
	//log.Printf("Outgoing headers: newrelic=%s traceparent=%s tracestate=%s",
	//	c.Get("newrelic"), c.Get("traceparent"), c.Get("tracestate"))
	utils.Logger.Debug(requestID, "Incoming request headers: %v", c.GetReqHeaders())
	utils.Logger.Debug(requestID, "Outgoing request headers: %v", req.Header)
	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if externalSeg != nil {
		defer externalSeg.End()
	}
	if err != nil {
		utils.Logger.Error(requestID, "error calling nisansala traced: %v", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "error", "error": err.Error()})
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.ReadAll(resp.Body)
	downstreamDuration := time.Since(start)

	// Attach some attributes to the transaction to show durations
	txn.AddAttribute("internal.duration_ms", internalDuration.Milliseconds())
	txn.AddAttribute("downstream.duration_ms", downstreamDuration.Milliseconds())

	utils.Logger.Info(requestID, "TracedHandler - end (internal %s, downstream %s)", r.String(), downstreamDuration.String())

	return c.JSON(fiber.Map{
		"status":              "OK",
		"internal_duration":   r.Milliseconds(),
		"downstream_duration": downstreamDuration.Milliseconds(),
	})
}
