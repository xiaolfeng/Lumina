package logic

import (
	"bytes"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/xiaolfeng/Lumina/internal/entity"
)

func TestWebAuthnCredentialRoundTrip(t *testing.T) {
	aaguid := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	original := &webauthn.Credential{
		ID:                []byte("credential-id"),
		PublicKey:         []byte("public-key"),
		AttestationType:   "basic_surrogate",
		AttestationFormat: "packed",
		Transport:         []protocol.AuthenticatorTransport{protocol.Internal, protocol.Hybrid},
		Flags:             webauthn.NewCredentialFlags(protocol.FlagUserPresent | protocol.FlagUserVerified | protocol.FlagBackupEligible | protocol.FlagBackupState),
		Authenticator: webauthn.Authenticator{
			AAGUID:       aaguid,
			SignCount:    42,
			CloneWarning: true,
			Attachment:   protocol.Platform,
		},
		Attestation: webauthn.CredentialAttestation{
			ClientDataJSON:     []byte("client-data"),
			ClientDataHash:     []byte("client-hash"),
			AuthenticatorData:  []byte("authenticator-data"),
			PublicKeyAlgorithm: -7,
			Object:             []byte("attestation-object"),
		},
	}

	stored, err := newBiometricCredentialEntity(original, "MacBook Pro")
	if err != nil {
		t.Fatalf("newBiometricCredentialEntity() error = %v", err)
	}
	if stored.AAGUID != "00010203-0405-0607-0809-0a0b0c0d0e0f" {
		t.Fatalf("AAGUID = %q", stored.AAGUID)
	}

	restored, err := entityToWebAuthnCredential(stored)
	if err != nil {
		t.Fatalf("entityToWebAuthnCredential() error = %v", err)
	}
	if !bytes.Equal(restored.ID, original.ID) || !bytes.Equal(restored.PublicKey, original.PublicKey) {
		t.Fatal("credential identity fields did not round-trip")
	}
	if restored.Flags.ProtocolValue() != original.Flags.ProtocolValue() {
		t.Fatalf("flags = %08b, want %08b", restored.Flags.ProtocolValue(), original.Flags.ProtocolValue())
	}
	if restored.Authenticator.CloneWarning != original.Authenticator.CloneWarning {
		t.Fatal("clone warning did not round-trip")
	}
	if restored.Authenticator.Attachment != original.Authenticator.Attachment {
		t.Fatal("authenticator attachment did not round-trip")
	}
	if !bytes.Equal(restored.Attestation.Object, original.Attestation.Object) {
		t.Fatal("attestation object did not round-trip")
	}
}

func TestEntityToWebAuthnCredentialLegacyFallback(t *testing.T) {
	legacy := &entity.BiometricCredential{
		CredentialID:   []byte("legacy-id"),
		PublicKey:      []byte("legacy-key"),
		AAGUID:         "00000000-0000-0000-0000-000000000000",
		SignCount:      7,
		TransportTypes: "internal,hybrid",
		LastUsedAt:     timePtr(time.Unix(1, 0)),
	}

	restored, err := entityToWebAuthnCredential(legacy)
	if err != nil {
		t.Fatalf("entityToWebAuthnCredential() error = %v", err)
	}
	if len(restored.Authenticator.AAGUID) != 16 {
		t.Fatalf("AAGUID length = %d, want 16", len(restored.Authenticator.AAGUID))
	}
	if restored.Authenticator.SignCount != 7 {
		t.Fatalf("sign count = %d, want 7", restored.Authenticator.SignCount)
	}
	if len(restored.Transport) != 2 {
		t.Fatalf("transport count = %d, want 2", len(restored.Transport))
	}
}

func TestHydrateLegacyCredentialFlags(t *testing.T) {
	user := NewLuminaWebAuthnUser("owner", "owner@example.com", []webauthn.Credential{
		{ID: []byte("other")},
		{ID: []byte("legacy")},
	})
	flags := protocol.FlagUserPresent | protocol.FlagUserVerified | protocol.FlagBackupEligible | protocol.FlagBackupState

	hydrateLegacyCredentialFlags(user, []byte("legacy"), flags)

	if got := user.credentials[0].Flags.ProtocolValue(); got != 0 {
		t.Fatalf("unrelated credential flags = %08b, want 0", got)
	}
	if got := user.credentials[1].Flags.ProtocolValue(); got != flags {
		t.Fatalf("legacy credential flags = %08b, want %08b", got, flags)
	}
}

func TestParseWebAuthnOrigins(t *testing.T) {
	got := parseWebAuthnOrigins(" https://lumina.example.com, http://localhost:3000 ,, ")
	want := []string{"https://lumina.example.com", "http://localhost:3000"}
	if len(got) != len(want) {
		t.Fatalf("origin count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("origin[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestChallengeTTL(t *testing.T) {
	if got, want := challengeTTL(300000), 330*time.Second; got != want {
		t.Fatalf("challengeTTL() = %s, want %s", got, want)
	}
}

func timePtr(value time.Time) *time.Time {
	return &value
}
