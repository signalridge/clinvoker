package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func setEventStreamHeaders(ctx huma.Context) {
	ctx.SetStatus(http.StatusOK)
	ctx.SetHeader("Content-Type", "text/event-stream")
	ctx.SetHeader("Cache-Control", "no-cache")
	ctx.SetHeader("Connection", "keep-alive")
	ctx.SetHeader("X-Accel-Buffering", "no")
}

func writeSSE(ctx huma.Context, data []byte) error {
	writer := ctx.BodyWriter()
	if _, err := writer.Write(data); err != nil {
		return err
	}
	if flusher, ok := writer.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}

func writeSSEJSON(ctx huma.Context, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return writeSSE(ctx, append(append([]byte("data: "), data...), []byte("\n\n")...))
}

func writeSSEEvent(ctx huma.Context, event string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	message := make([]byte, 0, len(event)+len(data)+16)
	message = append(message, "event: "...)
	message = append(message, event...)
	message = append(message, '\n')
	message = append(message, "data: "...)
	message = append(message, data...)
	message = append(message, '\n', '\n')
	return writeSSE(ctx, message)
}
