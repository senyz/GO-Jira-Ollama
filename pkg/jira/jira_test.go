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
		expectedCount  int
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
							"status": {"name": "In Progress"},
							"resolution": {"name": "Done"}
						}
					},
					{
						"key": "TEST-2",
						"fields": {
							"summary": "Test task 2",
							"description": "Description 2",
							"status": {"name": "To Do"},
							"resolution": null
						}
					}
				]
			}`,
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  2,
		},
		{
			name:           "API returns error status",
			projectKey:     "TEST",
			mockResponse:   `{"errorMessages":["Unauthorized"]}`,
			mockStatusCode: http.StatusUnauthorized,
			expectError:    true,
			expectedCount:  0,
		},
		{
			name:           "Invalid JSON response",
			projectKey:     "TEST",
			mockResponse:   `invalid json`,
			mockStatusCode: http.StatusOK,
			expectError:    true,
			expectedCount:  0,
		},
		{
			name:           "No issues in response",
			projectKey:     "TEST",
			mockResponse:   `{"issues":[]}`,
			mockStatusCode: http.StatusOK,
			expectError:    true,
			expectedCount:  0,
		},
		{
			name:       "Malformed issues structure",
			projectKey: "TEST",
			mockResponse: `{
				"issues": [
					{
						"key": "TEST-1",
						"fields": {
							"summary": "Test task",
							"status": {"name": "Open"}
						}
					}
				]
			}`,
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the request URL contains the project key
				expectedURL := "/rest/api/2/search?jql=project=" + tt.projectKey + "+AND+created>=startOfDay()"
				if r.URL.String() != expectedURL {
					t.Errorf("expected URL %s, got %s", expectedURL, r.URL.String())
				}

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
			if len(tasks) != tt.expectedCount {
				t.Errorf("expected %d tasks, got %d", tt.expectedCount, len(tasks))
				return
			}

			// Check task structure for successful cases
			if tt.expectedCount > 0 {
				task := tasks[0]
				if task.Key == "" {
					t.Error("task Key should not be empty")
				}
				if task.Fields.Summary == "" {
					t.Error("task Fields.Summary should not be empty")
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
		if r.Method == "POST" && r.Header.Get("Content-Type") != "application/json" {
			t.Error("Content-Type header not set correctly for POST with body")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"issues":[]}`))
	}))
	defer server.Close()

	// Test GET request with token
	resp, err := makeRequest("GET", server.URL, "test-token", nil)
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

func TestJiraTaskStructure(t *testing.T) {
	// Test specific JiraTask structure parsing
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"issues": [
				{
					"key": "TEST-123",
					"fields": {
						"summary": "Test Summary",
						"description": "Test Description",
						"status": {"name": "In Progress"},
						"resolution": {"name": "Fixed"}
					}
				}
			]
		}`))
	}))
	defer server.Close()

	tasks, err := GetJiraTask(server.URL, "test-token", "TEST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task.Key != "TEST-123" {
		t.Errorf("expected key TEST-123, got %s", task.Key)
	}
	if task.Fields.Summary != "Test Summary" {
		t.Errorf("expected summary 'Test Summary', got '%s'", task.Fields.Summary)
	}
	if task.Fields.Description != "Test Description" {
		t.Errorf("expected description 'Test Description', got '%s'", task.Fields.Description)
	}
	if task.Fields.Status.Name != "In Progress" {
		t.Errorf("expected status 'In Progress', got '%s'", task.Fields.Status.Name)
	}
	if task.Fields.Resolution.Name != "Fixed" {
		t.Errorf("expected resolution 'Fixed', got '%s'", task.Fields.Resolution.Name)
	}
}
