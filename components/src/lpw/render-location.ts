export interface LpwRenderLocation {
  jsonPath: string
  idPath: string[]
}

export function rootLocation(): LpwRenderLocation {
  return { jsonPath: '/content', idPath: [] }
}

export function childLocation(
  parent: LpwRenderLocation,
  index: number,
  id: string,
): LpwRenderLocation {
  return {
    jsonPath: `${parent.jsonPath}/${index}`,
    idPath: [...parent.idPath, id],
  }
}

export function formatLocation(loc: LpwRenderLocation): string {
  return loc.idPath.length > 0
    ? `${loc.jsonPath}（${loc.idPath.join(' › ')}）`
    : loc.jsonPath
}
