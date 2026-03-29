// Copyright 2026 Meridian OSS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0

package retry

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDoSucceedsSecondAttempt(t *testing.T) {
	ctx := context.Background()
	n := 0
	err := Do(ctx, 3, 0, func(int) error {
		n++
		if n < 2 {
			return errors.New("fail")
		}
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 2, n)
}
