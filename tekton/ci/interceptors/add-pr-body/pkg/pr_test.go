package pkg

import (
	"bytes"
	"context"
	"github.com/google/go-cmp/cmp"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1beta1"
	"google.golang.org/grpc/codes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInterceptor_Process(t *testing.T) {
	for _, tc := range []struct{
		name string
		req triggersv1.InterceptorRequest
		want triggersv1.InterceptorResponse
	}{{
		name: "empty extensions",
		req: triggersv1.InterceptorRequest{
			Extensions: map[string]interface{}{},
		},
		want: triggersv1.InterceptorResponse{
			Extensions: nil,
			Continue:   false,
			Status:     triggersv1.Status{
				Code: codes.FailedPrecondition,
				Message: "no 'add-pr-body' found in the extensions",
			},
		},
	}, {
		name: "no add-pr-body in extensions",
		req: triggersv1.InterceptorRequest{
			Extensions: map[string]interface{}{
				"foo": "bar",
			},
		},
		want: triggersv1.InterceptorResponse{
			Extensions: nil,
			Continue:   false,
			Status:     triggersv1.Status{
				Code: codes.FailedPrecondition,
				Message: "no 'add-pr-body' found in the extensions",
			},
		},
	}, {
		name: "no-pull-request-url-found",
		req: triggersv1.InterceptorRequest{
			Extensions: map[string]interface{}{
				"add_pr_body": map[string]interface{}{
					"foo": "bar",
				},
			},
		},
		want: triggersv1.InterceptorResponse{
			Extensions: nil,
			Continue:   false,
			Status:     triggersv1.Status{
				Code: codes.FailedPrecondition,
				Message: "no 'pull-request-url' found in the extensions",
			},
		},
	}, {
		name: "pull-request-url not a string",
		req: triggersv1.InterceptorRequest{
			Extensions: map[string]interface{}{
				"add_pr_body": map[string]interface{}{
					"pull_request_url": 4000,
				},
			},
		},
		want: triggersv1.InterceptorResponse{
			Extensions: nil,
			Continue:   false,
			Status:     triggersv1.Status{
				Code: codes.FailedPrecondition,
				Message: "'pull-request-url' found, but not a string",
			},
		},
	}}{
		t.Run(tc.name, func(t *testing.T) {
			i := Interceptor{}
			got := i.Process(context.Background(), &tc.req)
			if diff := cmp.Diff(&tc.want, got); diff != "" {
				t.Fatalf("-want/+got: %s", diff)
			}
		})
	}
}


type requestOption func(*http.Request)

// creates a GitHub hook type request - no secret is provided in testing.
func createRequest(method, url, event, token string, body []byte, opts ...requestOption) *http.Request {
	req := httptest.NewRequest(method, url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Github-Event", event)
	req.Header.Set("X-Github-Delivery", "testing-123")
	if token != "" {
		req.Header.Add("Authorization", "token "+token)
	}
	for _, o := range opts {
		o(req)
	}
	return req
}