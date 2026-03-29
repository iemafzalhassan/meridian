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

package retry

import (
	"context"
	"fmt"
	"time"
)

// Do runs fn up to attempts times (minimum 1). It respects ctx cancellation between attempts.
func Do(ctx context.Context, attempts int, backoff time.Duration, fn func(attempt int) error) error {
	if attempts < 1 {
		attempts = 1
	}
	var last error
	for i := 0; i < attempts; i++ {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("retry: %w", err)
		}
		last = fn(i)
		if last == nil {
			return nil
		}
		if i < attempts-1 && backoff > 0 {
			select {
			case <-ctx.Done():
				return fmt.Errorf("retry: %w", ctx.Err())
			case <-time.After(backoff):
			}
		}
	}
	return last
}
