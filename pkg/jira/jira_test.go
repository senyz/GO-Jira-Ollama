package jira

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetJiraTask(t *testing.T) {
	tests := []struct {
		name           string
		projectKey     string
		mockResponse   string
		mockStatusCode int
		expectError    bool
		expectedTasks  []map[string]interface{}
	}{
		{
			name:       "Successful response with tasks",
			projectKey: "TEST",
			mockResponse: `{
				"issues": [
					{
						"key": "TEST-1",
						"fields": {
							"summary": "Test task 1",
							"description": "Description 1",
							"resolution": {"name": "Done"}
						}
					},
					{
						"key": "TEST-2",
						"fields": {
							"summary": "Test task 2",
							"description": "Description 2",
							"resolution": null
						}
					}
				]
			}`,
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedTasks: []map[string]interface{}{
				{
					"key":         "TEST-1",
					"summary":     "Test task 1",
					"description": "Description 1",
					"resolution":  "Done",
				},
				{
					"key":         "TEST-2",
					"summary":     "Test task 2",
					"description": "Description 2",
					"resolution":  "",
				},
			},
		},
		{
			name:           "API returns error status",
			projectKey:     "TEST",
			mockResponse:   `{"errorMessages":["Unauthorized"]}`,
			mockStatusCode: http.StatusUnauthorized,
			expectError:    true,
		},
		{
			name:           "Invalid JSON response",
			projectKey:     "TEST",
			mockResponse:   `invalid json`,
			mockStatusCode: http.StatusOK,
			expectError:    true,
		},
		{
			name:           "No issues in response",
			projectKey:     "TEST",
			mockResponse:   `{"issues":[]}`,
			mockStatusCode: http.StatusOK,
			expectError:    true,
		},
		{
			name:       "Malformed issues structure",
			projectKey: "TEST",
			mockResponse: `{
				"issues": [
					{
						"key": "TEST-1"
					}
				]
			}`,
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedTasks:  []map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.mockStatusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			// Call function
			tasks, err := GetJiraTask(server.URL, "test-token", tt.projectKey)

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check tasks count
			if len(tasks) != len(tt.expectedTasks) {
				t.Errorf("expected %d tasks, got %d", len(tt.expectedTasks), len(tasks))
				return
			}

			// Check each task
			for i, expectedTask := range tt.expectedTasks {
				actualTask := tasks[i]
				for key, expectedValue := range expectedTask {
					if actualValue, ok := actualTask[key]; !ok || actualValue != expectedValue {
						t.Errorf("task %d field %s: expected %v, got %v", i, key, expectedValue, actualValue)
					}
				}
			}
		})
	}
}

func TestGetJiraTaskRequestFailure(t *testing.T) {
	// Test with invalid URL to trigger request error
	_, err := GetJiraTask("http://invalid-url-that-does-not-exist.local", "test-token", "TEST")
	if err == nil {
		t.Error("expected error for invalid URL but got none")
	}
}

func TestMakeRequest(t *testing.T) {
	// Test successful request with token and body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("Authorization header not set correctly")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("Content-Type header not set correctly")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	resp, err := makeRequest("POST", server.URL, "test-token", nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Test request creation error
	_, err = makeRequest("GET", "http://invalid-url-that-does-not-exist.local", "", nil)
	if err == nil {
		t.Error("expected error for invalid URL but got none")
	}
}
