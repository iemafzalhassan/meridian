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

package qdrantutil

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// HostPort parses a Qdrant gRPC URL (for example "http://localhost:6334") into host and port.
func HostPort(raw string) (host string, port int, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "localhost", 6334, nil
	}
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", 0, fmt.Errorf("qdrantutil: parse url: %w", err)
	}
	host = u.Hostname()
	if host == "" {
		host = "localhost"
	}
	portStr := u.Port()
	if portStr != "" {
		p, err := strconv.Atoi(portStr)
		if err != nil {
			return "", 0, fmt.Errorf("qdrantutil: parse port: %w", err)
		}
		return host, p, nil
	}
	switch strings.ToLower(u.Scheme) {
	case "https":
		return host, 6334, nil
	default:
		return host, 6334, nil
	}
}
