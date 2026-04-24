package keyring_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestFakeKeyring_GetMissing(t *testing.T) {
	t.Parallel()

	kr := testhelpers.NewFakeKeyring()
	_, err := kr.Get("svc", "user")
	require.Error(t, err)
}

func TestFakeKeyring_SetAndGet(t *testing.T) {
	t.Parallel()

	kr := testhelpers.NewFakeKeyring()
	require.NoError(t, kr.Set("svc", "user", "secret"))
	got, err := kr.Get("svc", "user")
	require.NoError(t, err)
	assert.Equal(t, "secret", got)
}

func TestFakeKeyring_Delete(t *testing.T) {
	t.Parallel()

	kr := testhelpers.NewFakeKeyring()
	require.NoError(t, kr.Set("svc", "user", "secret"))
	require.NoError(t, kr.Delete("svc", "user"))
	_, err := kr.Get("svc", "user")
	require.Error(t, err)
}

func TestFakeKeyring_GetErr(t *testing.T) {
	t.Parallel()

	injected := errors.New("boom get")
	kr := testhelpers.NewFakeKeyring()
	kr.GetErr = injected
	_, err := kr.Get("svc", "user")
	require.ErrorIs(t, err, injected)
}

func TestFakeKeyring_SetErr(t *testing.T) {
	t.Parallel()

	injected := errors.New("boom set")
	kr := testhelpers.NewFakeKeyring()
	kr.SetErr = injected
	err := kr.Set("svc", "user", "pw")
	require.ErrorIs(t, err, injected)
}

func TestFakeKeyring_DeleteErr(t *testing.T) {
	t.Parallel()

	injected := errors.New("boom del")
	kr := testhelpers.NewFakeKeyring()
	kr.DelErr = injected
	err := kr.Delete("svc", "user")
	require.ErrorIs(t, err, injected)
}

func TestFakeKeyring_Concurrent(t *testing.T) {
	t.Parallel()

	kr := testhelpers.NewFakeKeyring()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			user := fmt.Sprintf("u%d", i)
			_ = kr.Set("svc", user, fmt.Sprintf("pw%d", i))
			_, _ = kr.Get("svc", user)
			_ = kr.Delete("svc", user)
		}(i)
	}
	wg.Wait()
}
