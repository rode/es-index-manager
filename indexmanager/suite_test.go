// Copyright 2021 The Rode Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package indexmanager_test

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rode/es-index-manager/indexmanager"

	"github.com/brianvoe/gofakeit/v6"
	"go.uber.org/zap"
)

var logger = zap.NewNop()
var fake = gofakeit.New(0)

func TestIndexManagerPackage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IndexManager Suite")
}

type transportAction = func(req *http.Request) (*http.Response, error)
type mockEsTransport struct {
	receivedHttpRequests  []*http.Request
	preparedHttpResponses []*http.Response
	actions               []transportAction
}

func (m *mockEsTransport) Perform(req *http.Request) (*http.Response, error) {
	m.receivedHttpRequests = append(m.receivedHttpRequests, req)

	// if we have an action, return its result
	if len(m.actions) != 0 {
		action := m.actions[0]
		if action != nil {
			m.actions = append(m.actions[:0], m.actions[1:]...)
			return action(req)
		}
	}

	// if we have a prepared response, send it
	if len(m.preparedHttpResponses) != 0 {
		res := m.preparedHttpResponses[0]
		m.preparedHttpResponses = append(m.preparedHttpResponses[:0], m.preparedHttpResponses[1:]...)

		return res, nil
	}

	// return nil if we don't know what to do
	return nil, nil
}

func createRandomMapping() *VersionedMapping {
	return &VersionedMapping{
		Version: fake.Word(),
		Mappings: map[string]interface{}{
			fake.Word(): fake.Word(),
		},
	}
}

func createIndexOrAliasName(parts ...string) string {
	withPrefix := append([]string{"grafeas"}, parts...)

	return strings.Join(withPrefix, "-")
}

func createESBody(value interface{}) io.ReadCloser {
	responseBody, err := json.Marshal(value)
	Expect(err).To(BeNil())

	return ioutil.NopCloser(bytes.NewReader(responseBody))
}

func readRequestBody(request *http.Request, target interface{}) {
	rawBody, err := ioutil.ReadAll(request.Body)
	Expect(err).ToNot(HaveOccurred())

	Expect(json.Unmarshal(rawBody, target)).ToNot(HaveOccurred())
}

func createEsErrorResponse(errorType string) io.ReadCloser {
	return createESBody(map[string]interface{}{
		"error": map[string]interface{}{
			"type": errorType,
		},
	})
}
