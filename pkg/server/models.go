package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func (s *Server) Models(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	oauthToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if oauthToken == "" {
		var err error
		oauthToken, err = s.auth.GetToken(r.Context())
		if err != nil {
			respJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	ctx := r.Context()
	token, err := s.client.TokenWishCache(ctx, oauthToken)
	if err != nil {
		respJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models, err := s.client.Models(ctx, token)
	if err != nil {
		respJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	modelsResponse := ModelsResponse{
		Models: make([]Model, 0, len(models)),
	}
	for _, model := range models {
		modelsResponse.Models = append(modelsResponse.Models, Model{
			CreatedAt: time.Now().Unix(),
			ID:        model.ID,
			Object:    model.Object,
			OwnedBy:   model.Vendor,
			Permission: []Permission{
				{
					AllowCreateEngine: model.ModelPickerEnabled,
					AllowSampling:     model.ModelPickerEnabled,
					AllowLogprobs:     model.ModelPickerEnabled,
				},
			},
			Root:   model.ID,
			Parent: model.ID,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(modelsResponse); err != nil {
		respJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

// Model struct represents an OpenAPI model.
type Model struct {
	CreatedAt  int64        `json:"created"`
	ID         string       `json:"id"`
	Object     string       `json:"object"`
	OwnedBy    string       `json:"owned_by"`
	Permission []Permission `json:"permission"`
	Root       string       `json:"root"`
	Parent     string       `json:"parent"`
}

// Permission struct represents an OpenAPI permission.
type Permission struct {
	CreatedAt          int64       `json:"created"`
	ID                 string      `json:"id"`
	Object             string      `json:"object"`
	AllowCreateEngine  bool        `json:"allow_create_engine"`
	AllowSampling      bool        `json:"allow_sampling"`
	AllowLogprobs      bool        `json:"allow_logprobs"`
	AllowSearchIndices bool        `json:"allow_search_indices"`
	AllowView          bool        `json:"allow_view"`
	AllowFineTuning    bool        `json:"allow_fine_tuning"`
	Organization       string      `json:"organization"`
	Group              interface{} `json:"group"`
	IsBlocking         bool        `json:"is_blocking"`
}

type ModelsResponse struct {
	Models []Model `json:"data"`
}
