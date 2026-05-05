// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const apiBaseURL = "https://api.productive.io/api/v2"

// Client is the Productive.io API client used by all resources and data sources.
type Client struct {
	http           *http.Client
	token          string
	organizationID string
}

func NewClient(token, organizationID string) *Client {
	return &Client{
		http:           &http.Client{},
		token:          token,
		organizationID: organizationID,
	}
}

// --- JSON:API envelope types ---

type apiError struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

type apiErrorResponse struct {
	Errors []apiError `json:"errors"`
}

// personWriteAttrs are the fields sent to the API on Create and Update.
// Optional string fields are always sent (empty string clears the field).
// Optional integer fields use pointers so nil omits them from the payload.
type personWriteAttrs struct {
	FirstName                   string  `json:"first_name"`
	LastName                    string  `json:"last_name"`
	Email                       string  `json:"email"`
	Nickname                    string  `json:"nickname,omitempty"`
	Title                       string  `json:"title,omitempty"`
	TagList                     string  `json:"tag_list,omitempty"`
	RoleID                      *int64  `json:"role_id,omitempty"`
	CompanyID                   *int64  `json:"company_id,omitempty"`
	ManagerID                   *int64  `json:"manager_id,omitempty"`
	SubsidiaryID                *int64  `json:"subsidiary_id,omitempty"`
	CustomRoleID                *int64  `json:"custom_role_id,omitempty"`
	TimeTrackingPolicyID        *int64  `json:"time_tracking_policy_id,omitempty"`
	TimesheetSubmissionDisabled *bool   `json:"timesheet_submission_disabled,omitempty"`
	Virtual                     *bool   `json:"virtual,omitempty"`
}

// PersonAttributes are the fields returned by the API in read responses.
type PersonAttributes struct {
	FirstName                   string  `json:"first_name"`
	LastName                    string  `json:"last_name"`
	Email                       string  `json:"email"`
	Nickname                    string  `json:"nickname"`
	Title                       string  `json:"title"`
	TagList                     []string `json:"tag_list"`
	RoleID                      *int64  `json:"role_id"`
	CompanyID                   *int64  `json:"company_id"`
	ManagerID                   *int64  `json:"manager_id"`
	SubsidiaryID                *int64  `json:"subsidiary_id"`
	CustomRoleID                *int64  `json:"custom_role_id"`
	TimeTrackingPolicyID        *int64  `json:"time_tracking_policy_id"`
	TimesheetSubmissionDisabled bool    `json:"timesheet_submission_disabled"`
	Virtual                     bool    `json:"virtual"`
	Status                      int64   `json:"status"`
	CreatedAt                   string  `json:"created_at"`
	ArchivedAt                  *string `json:"archived_at"`
	InvitedAt                   *string `json:"invited_at"`
	LastSeenAt                  *string `json:"last_seen_at"`
	IsUser                      bool    `json:"is_user"`
	AvatarURL                   string  `json:"avatar_url"`
	TwoFactorAuth               bool    `json:"two_factor_auth"`
}

// PersonData is the JSON:API data object returned by the API.
type PersonData struct {
	Type       string           `json:"type"`
	ID         string           `json:"id"`
	Attributes PersonAttributes `json:"attributes"`
}

type personWriteData struct {
	Type       string           `json:"type"`
	ID         string           `json:"id,omitempty"`
	Attributes personWriteAttrs `json:"attributes"`
}

type personWriteRequest struct {
	Data personWriteData `json:"data"`
}

type personReadResponse struct {
	Data PersonData `json:"data"`
}

// --- HTTP helpers ---

func (c *Client) newRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, apiBaseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Accept", "application/vnd.api+json")
	req.Header.Set("X-Auth-Token", c.token)
	req.Header.Set("X-Organization-Id", c.organizationID)

	return req, nil
}

func (c *Client) do(req *http.Request, out any) (int, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr apiErrorResponse
		if jsonErr := json.Unmarshal(body, &apiErr); jsonErr == nil && len(apiErr.Errors) > 0 {
			return resp.StatusCode, fmt.Errorf("API error %d: %s — %s",
				resp.StatusCode, apiErr.Errors[0].Title, apiErr.Errors[0].Detail)
		}
		return resp.StatusCode, fmt.Errorf("unexpected HTTP %d: %s", resp.StatusCode, string(body))
	}

	if out != nil && len(body) > 0 {
		if err := json.Unmarshal(body, out); err != nil {
			return resp.StatusCode, fmt.Errorf("decoding response: %w", err)
		}
	}

	return resp.StatusCode, nil
}

// --- Person API methods ---

func (c *Client) CreatePerson(ctx context.Context, attrs personWriteAttrs) (*PersonData, error) {
	payload := personWriteRequest{
		Data: personWriteData{Type: "people", Attributes: attrs},
	}
	req, err := c.newRequest(ctx, http.MethodPost, "/people", payload)
	if err != nil {
		return nil, err
	}
	var out personReadResponse
	if _, err := c.do(req, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// GetPerson returns nil, nil when the person is not found (HTTP 404).
func (c *Client) GetPerson(ctx context.Context, id string) (*PersonData, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/people/"+id, nil)
	if err != nil {
		return nil, err
	}
	var out personReadResponse
	status, err := c.do(req, &out)
	if status == http.StatusNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out.Data, nil
}

func (c *Client) UpdatePerson(ctx context.Context, id string, attrs personWriteAttrs) (*PersonData, error) {
	payload := personWriteRequest{
		Data: personWriteData{Type: "people", ID: id, Attributes: attrs},
	}
	req, err := c.newRequest(ctx, http.MethodPatch, "/people/"+id, payload)
	if err != nil {
		return nil, err
	}
	var out personReadResponse
	if _, err := c.do(req, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// ArchivePerson soft-deletes a person. Productive.io has no hard DELETE.
func (c *Client) ArchivePerson(ctx context.Context, id string) error {
	req, err := c.newRequest(ctx, http.MethodPatch, "/people/"+id+"/archive", nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}
