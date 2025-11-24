package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"linuxdo-relay/internal/models"
	"linuxdo-relay/internal/relay"
)

// RegisterRelayRoutes registers provider-style relay endpoints that proxy
// transparently to new-api without any format conversion.
func RegisterRelayRoutes(r *gin.RouterGroup, app *AppContext) {
	client := relay.NewProxyClient()

	// OpenAI-compatible chat completions entrypoint.
	r.POST("/v1/chat/completions", func(c *gin.Context) {
		proxyToNewAPI(c, app, client, "/v1/chat/completions")
	})

	// Claude /v1/messages entrypoint.
	r.POST("/v1/messages", func(c *gin.Context) {
		proxyToNewAPI(c, app, client, "/v1/messages")
	})

	// Gemini generateContent entrypoint.
	r.POST("/v1beta/models/*path", func(c *gin.Context) {
		proxyToNewAPI(c, app, client, "")
	})
}

// proxyToNewAPI reads the request body once, determines the model, selects an
// appropriate channel, chooses the upstream path based on model name, and
// transparently proxies the request/response to/from new-api.
func proxyToNewAPI(c *gin.Context, app *AppContext, client *relay.ProxyClient, fixedPath string) {
	// Read entire body; quota middleware already read-and-reset body earlier.
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	// Restore body for potential further use by Gin (not strictly required here).
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	// Determine model depending on path.
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}

	var model string
	if strings.HasPrefix(path, "/v1/chat/completions") || strings.HasPrefix(path, "/v1/messages") {
		var tmp struct {
			Model string `json:"model"`
		}
		if err := json.Unmarshal(body, &tmp); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		model = tmp.Model
	} else if strings.HasPrefix(path, "/v1beta/models/") {
		param := c.Param("path")
		model = extractGeminiModelName(param)
	}

	if model == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model is required"})
		return
	}

	// Select a channel that supports this model.
	ch, err := pickChannelForModel(app, model)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	// Determine upstream path based on model name if not fixed by route.
	upPath := fixedPath
	if upPath == "" {
		upPath = determineUpstreamPath(model)
	}

	if upPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported model"})
		return
	}

	targetURL := strings.TrimRight(ch.BaseURL, "/") + upPath
	if strings.HasPrefix(upPath, "/v1beta/models/") {
		// Preserve query string (e.g. alt=sse) for Gemini.
		if q := c.Request.URL.RawQuery; q != "" {
			targetURL = targetURL + "?" + q
		}
	}

	statusCode, err := client.ProxyRequest(c.Writer, c.Request, http.MethodPost, targetURL, ch.APIKey, body)
	if err != nil {
		// Network or upstream transport error before we got a valid response.
		recordAPILogFromContext(app, c, model, 0, "fail", "upstream request failed: "+err.Error())
		c.JSON(http.StatusBadGateway, gin.H{"error": "upstream request failed"})
		return
	}

	// Successful round-trip to upstream; record log with actual status code.
	status := "success"
	var errMsg string
	if statusCode < 200 || statusCode >= 300 {
		status = "fail"
		// We don't parse body here; HTTP status code is usually enough for debugging.
	}
	recordAPILogFromContext(app, c, model, statusCode, status, errMsg)
}

// pickChannelForModel selects an enabled channel whose models JSON list
// contains the requested model name.
func pickChannelForModel(app *AppContext, model string) (*models.Channel, error) {
	var channels []models.Channel
	if err := app.DB.Where("status = ?", models.ChannelStatusEn).Find(&channels).Error; err != nil {
		return nil, fmt.Errorf("no available channel")
	}

	for i := range channels {
		var ms []string
		if err := json.Unmarshal([]byte(channels[i].Models), &ms); err != nil {
			continue
		}
		for _, m := range ms {
			if m == model {
				return &channels[i], nil
			}
		}
	}

	return nil, fmt.Errorf("no channel supports model %s", model)
}

// determineUpstreamPath maps a model name to the appropriate new-api path.
func determineUpstreamPath(model string) string {
	lower := strings.ToLower(model)

	// Gemini models: gemini-*.
	if strings.HasPrefix(lower, "gemini-") {
		return "/v1beta/models/" + model + ":generateContent"
	}

	// Anthropic Claude models.
	if strings.HasPrefix(lower, "claude-") {
		return "/v1/messages"
	}

	// Default: treat as OpenAI chat completion model.
	return "/v1/chat/completions"
}

// extractGeminiModelName parses model name from a Gemini path like
// "models/gemini-1.5-pro:generateContent".
func extractGeminiModelName(path string) string {
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return ""
	}
	last := parts[len(parts)-1]
	if idx := strings.Index(last, ":"); idx >= 0 {
		return last[:idx]
	}
	return last
}
