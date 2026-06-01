import { createFileRoute, Outlet } from '@tanstack/react-router'

import { Navbar } from '#/components/Navbar'
import { Footer } from '#/components/Footer'

export const Route = createFileRoute('/_public')({
  component: PublicLayout,
})

function PublicLayout() {
  return (
    <div className="min-h-screen pt-16 md:pt-20">
      <Navbar />
      <Outlet />
      <Footer />
    </div>
  )
}
