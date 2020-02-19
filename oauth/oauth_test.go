package oauth

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/apiotrowski312/isOnline-utils-go/http_utils"
	"github.com/apiotrowski312/isOnline-utils-go/rest_errors"
	"github.com/stretchr/testify/assert"
)

func init() {
	http_utils.Client = &http_utils.MockClient{}
}

func TestOauthConstants(t *testing.T) {
	assert.EqualValues(t, headerXPublic, "X-Public")
	assert.EqualValues(t, headerXClientId, "X-Client-Id")
	assert.EqualValues(t, headerXCallerId, "X-Caller-Id")
	assert.EqualValues(t, paramAccessToken, "access_token")
}

func TestIsPublic(t *testing.T) {
	type args struct {
		request *http.Request
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"nil request",
			args{request: nil},
			true,
		},
		{
			"empty request",
			args{request: &http.Request{}},
			false,
		},
		{
			"request with truthy proper header",
			args{request: &http.Request{Header: http.Header{headerXPublic: []string{"true"}}}},
			true,
		},
		{
			"request with falsy proper header",
			args{request: &http.Request{Header: http.Header{headerXPublic: []string{"false"}}}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPublic(tt.args.request); got != tt.want {
				t.Errorf("IsPublic() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCallerId(t *testing.T) {
	type args struct {
		request *http.Request
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			"nil request",
			args{request: nil},
			0,
		},
		{
			"empty request",
			args{request: &http.Request{}},
			0,
		},
		{
			"request header is string intiger",
			args{request: &http.Request{Header: http.Header{headerXCallerId: []string{"123"}}}},
			123,
		},
		{
			"request header is not string intiger",
			args{request: &http.Request{Header: http.Header{headerXCallerId: []string{"NotInt"}}}},
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCallerId(tt.args.request); got != tt.want {
				t.Errorf("GetCallerId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetClientId(t *testing.T) {
	type args struct {
		request *http.Request
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			"nil request",
			args{request: nil},
			0,
		},
		{
			"empty request",
			args{request: &http.Request{}},
			0,
		},
		{
			"request header is string intiger",
			args{request: &http.Request{Header: http.Header{headerXClientId: []string{"123"}}}},
			123,
		},
		{
			"request header is not string intiger",
			args{request: &http.Request{Header: http.Header{headerXClientId: []string{"NotInt"}}}},
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetClientId(tt.args.request); got != tt.want {
				t.Errorf("GetClientId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cleanRequest(t *testing.T) {
	type args struct {
		request *http.Request
	}
	tests := []struct {
		name string
		args args
		want args
	}{
		{
			"nil request",
			args{request: nil},
			args{request: nil},
		},
		{
			"empty request",
			args{request: &http.Request{}},
			args{request: &http.Request{}},
		},
		{
			"request with one header to remove",
			args{request: &http.Request{Header: http.Header{headerXClientId: []string{"NotInt"}}}},
			args{request: &http.Request{}},
		},
		{
			"request with two headers to remove",
			args{request: &http.Request{Header: http.Header{headerXClientId: []string{"NotInt"}, headerXCallerId: []string{"NotInt"}}}},
			args{request: &http.Request{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanRequest(tt.args.request)
			if tt.args.request != nil && tt.args.request.Header.Get(headerXClientId) != tt.want.request.Header.Get(headerXClientId) {
				t.Errorf("GetClientId() headerXClientId  = %v, want %v", tt.args.request.Header.Get(headerXClientId), tt.want.request.Header.Get(headerXClientId))
			}
			if tt.args.request != nil && tt.args.request.Header.Get(headerXCallerId) != tt.want.request.Header.Get(headerXCallerId) {
				t.Errorf("GetClientId() headerXCallerId  = %v, want %v", tt.args.request.Header.Get(headerXCallerId), tt.want.request.Header.Get(headerXCallerId))
			}
		})
	}
}

func Test_getAccessToken(t *testing.T) {
	type args struct {
		accessTokenId string
	}

	accessToken200 := &accessToken{
		Id:       "123",
		UserId:   123,
		ClientId: 123,
	}

	tests := []struct {
		name  string
		args  args
		mock  func(*http.Request) (*http.Response, error)
		want  *accessToken
		want1 rest_errors.RestErr
	}{
		{
			"200 request",
			args{accessTokenId: "123"},
			func(*http.Request) (*http.Response, error) {
				jsonBytes, _ := json.Marshal(accessToken200)
				return &http.Response{
					StatusCode: 200,
					Body:       ioutil.NopCloser(bytes.NewReader(jsonBytes)),
				}, nil
			},
			accessToken200,
			nil,
		},
		{
			"not json 500 request",
			args{accessTokenId: "123"},
			func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 500,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte("not json"))),
				}, nil
			},
			nil,
			rest_errors.NewInternalServerError("invalid error interface when try get access token", errors.New("login error")),
		},
		{
			"not json 200 request",
			args{accessTokenId: "123"},
			func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte("not json"))),
				}, nil
			},
			nil,
			rest_errors.NewInternalServerError("error when trying to unmarshal access token response", errors.New("login error")),
		},
		{
			"Test error",
			args{accessTokenId: "123"},
			func(*http.Request) (*http.Response, error) {
				jsonBytes, _ := json.Marshal(rest_errors.NewBadRequestError("Error on purpose"))
				return &http.Response{
					StatusCode: 500,
					Body:       ioutil.NopCloser(bytes.NewReader(jsonBytes)),
				}, nil
			},
			nil,
			rest_errors.NewBadRequestError("Error on purpose"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			http_utils.GetDoFunc = tt.mock

			got, got1 := getAccessToken(tt.args.accessTokenId)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAccessToken() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("getAccessToken() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestAuthenticateRequest(t *testing.T) {
	type args struct {
		request *http.Request
	}
	tests := []struct {
		name string
		mock func(*http.Request) (*http.Response, error)
		args args
		want rest_errors.RestErr
	}{
		{
			"nil request",
			func(*http.Request) (*http.Response, error) { return nil, nil },
			args{nil},
			nil,
		},
		{
			"error request",
			func(*http.Request) (*http.Response, error) { return nil, errors.New("Error on purpose") },
			args{
				request: &http.Request{
					URL: &url.URL{
						RawQuery: "access_token=123",
					},
					Header: http.Header{},
				},
			},
			rest_errors.NewInternalServerError("invalid restclient response when trying to get access token", errors.New("login error")),
		},
		{
			"no access token request",
			func(*http.Request) (*http.Response, error) { return nil, nil },
			args{
				request: &http.Request{
					URL: &url.URL{RawQuery: "access_token="},
				},
			},
			nil,
		},
		{
			"proper request",
			func(*http.Request) (*http.Response, error) {
				jsonBytes, _ := json.Marshal(&accessToken{
					Id:       "123",
					UserId:   123,
					ClientId: 123,
				},
				)
				return &http.Response{
					StatusCode: 200,
					Body:       ioutil.NopCloser(bytes.NewReader(jsonBytes)),
				}, nil
			},
			args{
				request: &http.Request{
					URL: &url.URL{
						RawQuery: "access_token=123",
					},
					Header: http.Header{},
				},
			},
			nil,
		},
		{
			"404 request",
			func(*http.Request) (*http.Response, error) {
				jsonBytes, _ := json.Marshal(rest_errors.NewNotFoundError("Not found"))
				return &http.Response{
					StatusCode: 404,
					Body:       ioutil.NopCloser(bytes.NewReader(jsonBytes)),
				}, nil
			},
			args{
				request: &http.Request{
					URL: &url.URL{
						RawQuery: "access_token=123",
					},
					Header: http.Header{},
				},
			},
			nil,
		},
	}
	for _, tt := range tests {
		http_utils.GetDoFunc = tt.mock

		t.Run(tt.name, func(t *testing.T) {
			if got := AuthenticateRequest(tt.args.request); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AuthenticateRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
