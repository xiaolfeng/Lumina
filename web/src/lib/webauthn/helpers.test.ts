import { afterEach, describe, expect, it, vi } from 'vitest'
import { createCredential, credentialToJSON, getCredential } from './helpers'
import type {
  PublicKeyCredentialCreationOptionsJSON,
  PublicKeyCredentialRequestOptionsJSON,
} from './helpers'

const creationJSON: PublicKeyCredentialCreationOptionsJSON = {
  rp: { id: 'localhost', name: 'Lumina' },
  user: {
    id: 'bHVtaW5hLW93bmVy',
    name: 'owner',
    displayName: 'Owner',
  },
  challenge: 'AQIDBA',
  pubKeyCredParams: [{ type: 'public-key', alg: -7 }],
}

const requestJSON: PublicKeyCredentialRequestOptionsJSON = {
  challenge: 'AQIDBA',
  rpId: 'localhost',
  allowCredentials: [{ type: 'public-key', id: 'BQYHCA' }],
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('WebAuthn option wrapping', () => {
  it('wraps native creation parser output in publicKey', async () => {
    const parsed = { challenge: new Uint8Array([1, 2, 3, 4]) }
    const create = vi.fn().mockResolvedValue({ id: 'credential' })
    vi.stubGlobal('window', {
      PublicKeyCredential: {
        parseCreationOptionsFromJSON: vi.fn().mockReturnValue(parsed),
      },
    })
    vi.stubGlobal('navigator', { credentials: { create } })

    await createCredential(creationJSON)

    expect(create).toHaveBeenCalledWith({ publicKey: parsed })
  })

  it('wraps manually decoded creation options in publicKey', async () => {
    const create = vi.fn().mockResolvedValue({ id: 'credential' })
    vi.stubGlobal('window', { PublicKeyCredential: {} })
    vi.stubGlobal('navigator', { credentials: { create } })

    await createCredential(creationJSON)

    const call = create.mock.calls[0]?.[0] as CredentialCreationOptions
    expect(call).toEqual({ publicKey: expect.any(Object) })
    expect(call.publicKey?.challenge).toBeInstanceOf(ArrayBuffer)
    expect(call.publicKey?.user.id).toBeInstanceOf(ArrayBuffer)
  })

  it('wraps native request parser output in publicKey', async () => {
    const parsed = { challenge: new Uint8Array([1, 2, 3, 4]) }
    const get = vi.fn().mockResolvedValue({ id: 'credential' })
    vi.stubGlobal('window', {
      PublicKeyCredential: {
        parseRequestOptionsFromJSON: vi.fn().mockReturnValue(parsed),
      },
    })
    vi.stubGlobal('navigator', { credentials: { get } })

    await getCredential(requestJSON)

    expect(get).toHaveBeenCalledWith({ publicKey: parsed })
  })

  it('wraps manually decoded request options in publicKey', async () => {
    const get = vi.fn().mockResolvedValue({ id: 'credential' })
    vi.stubGlobal('window', { PublicKeyCredential: {} })
    vi.stubGlobal('navigator', { credentials: { get } })

    await getCredential(requestJSON)

    const call = get.mock.calls[0]?.[0] as CredentialRequestOptions
    expect(call).toEqual({ publicKey: expect.any(Object) })
    expect(call.publicKey?.challenge).toBeInstanceOf(ArrayBuffer)
    expect(call.publicKey?.allowCredentials?.[0]?.id).toBeInstanceOf(
      ArrayBuffer,
    )
  })
})

describe('WebAuthn credential serialization', () => {
  it('uses the native credential toJSON implementation when available', () => {
    const expected = {
      id: 'credential',
      response: { transports: ['internal'] },
    }
    const toJSON = vi.fn().mockReturnValue(expected)
    const credential = { toJSON } as unknown as PublicKeyCredential

    expect(credentialToJSON(credential)).toBe(expected)
    expect(toJSON).toHaveBeenCalledOnce()
  })

  it('preserves rawId, transports, and extension results in the fallback serializer', () => {
    const response = {
      attestationObject: new Uint8Array([1, 2]).buffer,
      clientDataJSON: new Uint8Array([3, 4]).buffer,
      getTransports: vi.fn().mockReturnValue(['internal', 'hybrid']),
    } as unknown as AuthenticatorAttestationResponse
    const credential = {
      id: 'credential',
      rawId: new Uint8Array([5, 6]).buffer,
      type: 'public-key',
      authenticatorAttachment: 'platform',
      response,
      getClientExtensionResults: vi
        .fn()
        .mockReturnValue({ credProps: { rk: true } }),
    } as unknown as PublicKeyCredential

    expect(credentialToJSON(credential)).toEqual({
      id: 'credential',
      rawId: 'BQY',
      type: 'public-key',
      authenticatorAttachment: 'platform',
      clientExtensionResults: { credProps: { rk: true } },
      response: {
        attestationObject: 'AQI',
        clientDataJSON: 'AwQ',
        transports: ['internal', 'hybrid'],
      },
    })
  })
})
