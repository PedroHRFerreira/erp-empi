import type { ILoginResponse } from '../../contracts/types'

export const loginResponseMock: ILoginResponse = {
  user: {
    id: 'admin-id',
    name: 'Administrador EMPI',
    cpf: '52998224725',
    type: 'admin',
    email: 'admin@empi.local',
    phone: '33987351922',
    markupPercent: 10,
    machineFeePercent: 0,
    installmentFeePercent: 0,
    address: '',
    notes: '',
    createdAt: '2026-06-16T00:00:00Z',
    updatedAt: '2026-06-16T00:00:00Z'
  },
  tokens: {
    accessToken: 'access-token',
    refreshToken: 'refresh-token'
  }
}
