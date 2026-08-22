package htcli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	hosttracker "github.com/HostTracker/hosttracker-sdk-go"
)

// Request is one call, as the CLI assembles it before handing it to the
// SDK.
type Request struct {
	Method string
	Path   string
	Query  url.Values
	Header http.Header
	Body   []byte
}

// Answer is a successful reply, read into memory.
type Answer struct {
	Status      int
	Header      http.Header
	Body        []byte
	ContentType string
}

// JSON reports whether the answer carries a JSON document.
func (a *Answer) JSON() bool {
	media := a.ContentType
	if i := strings.IndexByte(media, ';'); i >= 0 {
		media = media[:i]
	}
	media = strings.TrimSpace(strings.ToLower(media))
	return media == "" || strings.HasSuffix(media, "json")
}

// Do sends one request through the SDK. Everything on the wire, the
// bearer token, the idempotency key, the retry ladder and the mapping of
// a >= 400 answer onto *hosttracker.Error, is the SDK's doing.
func (o *Options) Do(ctx context.Context, req Request) (*Answer, error) {
	client, err := o.Client()
	if err != nil {
		return nil, err
	}
	target, err := resolveURL(client.BaseURL(), req.Path, req.Query)
	if err != nil {
		return nil, err
	}

	var body io.Reader
	if len(req.Body) > 0 {
		body = bytes.NewReader(req.Body)
	}
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, target, body)
	if err != nil {
		return nil, err
	}
	for name, values := range req.Header {
		for _, value := range values {
			httpReq.Header.Add(name, value)
		}
	}
	if len(req.Body) > 0 && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := client.HTTPDoer().Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &Answer{
		Status:      resp.StatusCode,
		Header:      resp.Header,
		Body:        raw,
		ContentType: resp.Header.Get("Content-Type"),
	}, nil
}

// resolveURL joins a path and its query onto the client's base URL. The
// base may carry a path prefix of its own, which a proxy in front of the
// API needs kept.
func resolveURL(base, path string, query url.Values) (string, error) {
	parsed, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	parsed.Path = strings.TrimSuffix(parsed.Path, "/") + "/" + strings.TrimPrefix(path, "/")
	if len(query) > 0 {
		parsed.RawQuery = query.Encode()
	}
	return parsed.String(), nil
}

// Emit prints one answer: the JSON document through the chosen format, a
// binary payload straight through, an empty answer not at all.
func (o *Options) Emit(answer *Answer) error {
	if answer.Status == http.StatusNoContent || len(bytes.TrimSpace(answer.Body)) == 0 {
		return nil
	}
	if !answer.JSON() {
		_, err := o.Out.Write(answer.Body)
		return err
	}
	return o.Printer().Print(json.RawMessage(answer.Body))
}

// collect walks every page of a paged answer and returns one envelope
// holding all the rows. The walk itself is the SDK's Paginate, so the
// cursor protocol lives in exactly one place.
func (o *Options) collect(ctx context.Context, req Request, cursorInBody bool) (*Answer, error) {
	var first *Answer
	rows := []json.RawMessage{}

	fetch := func(ctx context.Context, cursor *string) ([]json.RawMessage, *string, error) {
		page := req
		if cursor != nil {
			var err error
			if page, err = withCursor(req, *cursor, cursorInBody); err != nil {
				return nil, nil, err
			}
		}
		answer, err := o.Do(ctx, page)
		if err != nil {
			return nil, nil, err
		}
		if first == nil {
			first = answer
		}
		var envelope struct {
			Data       []json.RawMessage `json:"data"`
			NextCursor *string           `json:"nextCursor"`
		}
		if err := json.Unmarshal(answer.Body, &envelope); err != nil {
			return nil, nil, fmt.Errorf("--all needs a paged answer, and this one did not parse as one: %w", err)
		}
		return envelope.Data, envelope.NextCursor, nil
	}

	for row, err := range hosttracker.Paginate(ctx, fetch) {
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	if first == nil {
		return nil, nil
	}

	// Keep whatever else the envelope published (count, summary), and
	// close out the paging members: there is no next page left.
	envelope := map[string]json.RawMessage{}
	_ = json.Unmarshal(first.Body, &envelope)
	delete(envelope, "syncCursor")
	data, err := json.Marshal(rows)
	if err != nil {
		return nil, err
	}
	envelope["data"] = data
	envelope["hasMore"] = json.RawMessage("false")
	envelope["nextCursor"] = json.RawMessage("null")
	merged, err := json.Marshal(envelope)
	if err != nil {
		return nil, err
	}
	first.Body = merged
	return first, nil
}

// withCursor returns req addressed at the next page, putting the cursor
// where that form of the call carries it.
func withCursor(req Request, cursor string, inBody bool) (Request, error) {
	if !inBody {
		query := url.Values{}
		for name, values := range req.Query {
			query[name] = append([]string(nil), values...)
		}
		query.Set("cursor", cursor)
		req.Query = query
		return req, nil
	}
	document := map[string]any{}
	if len(bytes.TrimSpace(req.Body)) > 0 {
		if err := json.Unmarshal(req.Body, &document); err != nil {
			return req, fmt.Errorf("the query body is not a JSON object: %w", err)
		}
	}
	document["cursor"] = cursor
	raw, err := json.Marshal(document)
	if err != nil {
		return req, err
	}
	req.Body = raw
	return req, nil
}
