import { createFileRoute } from '@tanstack/react-router'

import { HeroSection } from '#/components/landing/hero-section'
import { FeaturesSection } from '#/components/landing/features-section'
import { TechSection } from '#/components/landing/tech-section'

export const Route = createFileRoute('/_public/')({ component: Home })

function Home() {
  return (
    <>
      <HeroSection />
      <FeaturesSection />
      <TechSection />
    </>
  )
}
