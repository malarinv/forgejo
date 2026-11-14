// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package webhook_test

import (
	"testing"

	"forgejo.org/modules/webhook"
)

// TestHookEventTypeEventPackage is a fork protecting test: the fork grafts
// `case HookEventPackage` into HookEventType.Event() so package webhooks
// announce X-Gitea-Event: package. Stock upstream v17 lacks that case
// (Event() returns "" for HookEventPackage) and the Flux receiver cannot
// match package pushes without it — this test fails on stock v17.
func TestHookEventTypeEventPackage(t *testing.T) {
	if got := webhook.HookEventPackage.Event(); got != "package" {
		t.Fatalf("HookEventPackage.Event() = %q, want %q", got, "package")
	}
	if got := webhook.HookEventPush.Event(); got != "push" {
		t.Fatalf("HookEventPush.Event() = %q, want %q", got, "push")
	}
	if got := webhook.HookEventRelease.Event(); got != "release" {
		t.Fatalf("HookEventRelease.Event() = %q, want %q", got, "release")
	}
}
