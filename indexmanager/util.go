package indexmanager

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

func decodeResponse(r io.ReadCloser, i interface{}) error {
	err := json.NewDecoder(r).Decode(i)
	if err != nil {
		return errors.New(fmt.Sprintf("error decoding elasticsearch response: %s", err))
	}

	return nil
}

func encodeRequest(body interface{}) (io.Reader, string) {
	b, err := json.Marshal(body)
	if err != nil {
		// we should know that `body` is a serializable struct before invoking `EncodeRequest`
		panic(err)
	}

	return bytes.NewReader(b), string(b)
}

func getErrorFromESResponse(res *esapi.Response, err error) error {
	if err != nil {
		return err
	}

	if res.IsError() {
		return fmt.Errorf("response error from ES: %d", res.StatusCode)
	}
	return nil
}

func parseIndexName(indexName string) *IndexName {
	// the index name is assumed to match one of the following types
	// $PREFIX-$VERSION-$DOCUMENT_KIND
	// $PREFIX-$VERSION-$INNER-$DOCUMENT_KIND

	parts := strings.Split(indexName, indexNamePartsDelimiter)
	documentKind := parts[len(parts)-1]
	name := &IndexName{
		DocumentKind: documentKind,
	}

	name.Version = parts[1]
	// TODO: length check
	name.Inner = strings.Join(parts[2:(len(parts)-1)], indexNamePartsDelimiter)

	return name
}
