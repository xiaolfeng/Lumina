import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import * as biometricApi from '#/lib/apis/biometric'
import * as userApi from '#/lib/apis/user'
import { createCredential, bufferToBase64url } from '#/lib/webauthn/helpers'
import type { RegisterFinishResponseWrapper } from '#/lib/models/response/biometric'
import type { BaseResponse } from '#/lib/models/response/common'
import type { BiometricCredentialListResponseWrapper } from '#/lib/models/response/user'

export function useBiometricCredentials() {
  return useQuery<BiometricCredentialListResponseWrapper, Error>({
    queryKey: ['biometric', 'credentials'],
    queryFn: userApi.getBiometricCredentials,
    staleTime: 30 * 1000,
  })
}

export function useRegisterBiometric() {
  const queryClient = useQueryClient()
  return useMutation<RegisterFinishResponseWrapper, Error, string>({
    mutationFn: async (deviceName: string) => {
      const startResp = await biometricApi.registerStart({ device_name: deviceName })
      const startData = startResp.data
      if (!startData) throw new Error('注册开始失败：未返回数据')

      const credential = await createCredential(startData.options)
      if (!credential) throw new Error('生物特征注册已取消')

      const credentialJSON = serializeRegistrationCredential(credential, deviceName)

      return biometricApi.registerFinish({
        session_token: startData.session_token,
        device_name: deviceName,
        credential: credentialJSON,
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['biometric', 'credentials'] })
      queryClient.invalidateQueries({ queryKey: ['biometric', 'availability'] })
      queryClient.invalidateQueries({ queryKey: ['user', 'current'] })
    },
  })
}

export function useDeleteBiometric() {
  const queryClient = useQueryClient()
  return useMutation<BaseResponse<null>, Error, string>({
    mutationFn: (id: string) => userApi.deleteBiometricCredential(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['biometric', 'credentials'] })
      queryClient.invalidateQueries({ queryKey: ['biometric', 'availability'] })
      queryClient.invalidateQueries({ queryKey: ['user', 'current'] })
    },
  })
}

function serializeRegistrationCredential(credential: PublicKeyCredential, deviceName: string): unknown {
  const response = credential.response as AuthenticatorAttestationResponse
  return {
    id: credential.id,
    rawId: credential.id,
    type: credential.type,
    response: {
      attestationObject: bufferToBase64url(response.attestationObject),
      clientDataJSON: bufferToBase64url(response.clientDataJSON),
    },
    authenticatorAttachment: credential.authenticatorAttachment,
    deviceName,
    clientExtensionResults: {},
  }
}
