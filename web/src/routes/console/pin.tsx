import { createFileRoute } from '@tanstack/react-router'
import { PinList } from '#/components/pin/pin-list'

export const Route = createFileRoute('/console/pin')({
  component: PinPage,
})

function PinPage() {
  return <PinList />
}
