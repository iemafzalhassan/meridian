// Copyright 2026 Meridian OSS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vectorstore

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/qdrant/go-client/qdrant"
)

// ProjectDocPayload builds a Qdrant-compatible payload map from a document.
func ProjectDocPayload(d ProjectDoc) (map[string]any, error) {
	b, err := json.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("vectorstore: marshal project doc: %w", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("vectorstore: project doc to map: %w", err)
	}
	return m, nil
}

// ProjectDocFromPayload recovers a ProjectDoc from Qdrant payload values.
func ProjectDocFromPayload(payload map[string]*qdrant.Value) (ProjectDoc, error) {
	if payload == nil {
		return ProjectDoc{}, fmt.Errorf("vectorstore: empty payload")
	}
	m := make(map[string]any, len(payload))
	for k, v := range payload {
		av, err := qdrantValueToAny(v)
		if err != nil {
			return ProjectDoc{}, fmt.Errorf("vectorstore: decode payload key %q: %w", k, err)
		}
		m[k] = av
	}
	b, err := json.Marshal(m)
	if err != nil {
		return ProjectDoc{}, fmt.Errorf("vectorstore: encode payload map: %w", err)
	}
	var doc ProjectDoc
	if err := json.Unmarshal(b, &doc); err != nil {
		return ProjectDoc{}, fmt.Errorf("vectorstore: payload to project doc: %w", err)
	}
	return doc, nil
}

func qdrantValueToAny(v *qdrant.Value) (any, error) {
	if v == nil {
		return nil, nil
	}
	switch k := v.GetKind().(type) {
	case *qdrant.Value_NullValue:
		return nil, nil
	case *qdrant.Value_DoubleValue:
		return k.DoubleValue, nil
	case *qdrant.Value_IntegerValue:
		return k.IntegerValue, nil
	case *qdrant.Value_StringValue:
		return k.StringValue, nil
	case *qdrant.Value_BoolValue:
		return k.BoolValue, nil
	case *qdrant.Value_StructValue:
		return structToMap(k.StructValue)
	case *qdrant.Value_ListValue:
		return listToSlice(k.ListValue)
	default:
		return nil, fmt.Errorf("unknown value kind %T", k)
	}
}

func structToMap(s *qdrant.Struct) (map[string]any, error) {
	if s == nil {
		return nil, nil
	}
	out := make(map[string]any, len(s.GetFields()))
	for key, val := range s.GetFields() {
		av, err := qdrantValueToAny(val)
		if err != nil {
			return nil, err
		}
		out[key] = av
	}
	return out, nil
}

func listToSlice(l *qdrant.ListValue) ([]any, error) {
	if l == nil {
		return nil, nil
	}
	out := make([]any, 0, len(l.GetValues()))
	for _, val := range l.GetValues() {
		av, err := qdrantValueToAny(val)
		if err != nil {
			return nil, err
		}
		out = append(out, av)
	}
	return out, nil
}

// StripJSONFences removes optional ```json fences from LLM output.
func StripJSONFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```JSON")
		s = strings.TrimPrefix(s, "```")
		if i := strings.LastIndex(s, "```"); i >= 0 {
			s = s[:i]
		}
	}
	return strings.TrimSpace(s)
}
