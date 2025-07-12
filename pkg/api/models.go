package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const defaultModelsURI = "https://api.githubcopilot.com/models"

func (c *Client) Models(ctx context.Context, token string) ([]Model, error) {
	response, err := c.doGet(ctx, token, defaultModelsURI)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return nil, statusError{
			StatusCode:   response.StatusCode,
			Status:       response.Status,
			ErrorMessage: string(body),
		}
	}

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(bytes))

	var models ModelsResponse
	if err := json.Unmarshal(bytes, &models); err != nil {
		fmt.Println(string(bytes))
		return nil, err
	}
	return models.Data, nil
}

type ModelsResponse struct {
	Data   []Model `json:"data"`
	Object string  `json:"object"`
}

type Model struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	Object             string            `json:"object"`
	Version            string            `json:"version"`
	Vendor             string            `json:"vendor"`
	Preview            bool              `json:"preview"`
	ModelPickerEnabled bool              `json:"model_picker_enabled"`
	Capabilities       ModelCapabilities `json:"capabilities"`
	Policy             *ModelPolicy      `json:"policy,omitempty"`
}

type ModelCapabilities struct {
	Family    string                `json:"family"`
	Object    string                `json:"object"`
	Type      string                `json:"type"`
	Tokenizer string                `json:"tokenizer"`
	Supports  ModelSupports         `json:"supports"`
	Limits    ModelCapabilityLimits `json:"limits"`
}

type ModelSupports struct {
	Streaming         bool `json:"streaming,omitempty"`
	ToolCalls         bool `json:"tool_calls,omitempty"`
	ParallelToolCalls bool `json:"parallel_tool_calls,omitempty"`
	Vision            bool `json:"vision,omitempty"`
	StructuredOutputs bool `json:"structured_outputs,omitempty"`
	Dimensions        bool `json:"dimensions,omitempty"`
}

type ModelCapabilityLimits struct {
	MaxContextWindowTokens int           `json:"max_context_window_tokens,omitempty"`
	MaxOutputTokens        int           `json:"max_output_tokens,omitempty"`
	MaxPromptTokens        int           `json:"max_prompt_tokens,omitempty"`
	MaxInputs              int           `json:"max_inputs,omitempty"`
	Vision                 *VisionLimits `json:"vision,omitempty"`
}

type VisionLimits struct {
	MaxPromptImageSize  int      `json:"max_prompt_image_size"`
	MaxPromptImages     int      `json:"max_prompt_images"`
	SupportedMediaTypes []string `json:"supported_media_types"`
}

type ModelPolicy struct {
	State string `json:"state"`
	Terms string `json:"terms"`
}
