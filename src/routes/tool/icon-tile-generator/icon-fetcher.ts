import type { Icon, IconSource } from './types';

const ICON_BASE_URLS: Record<IconSource, string> = {
  lucide: 'https://unpkg.com/lucide-static@latest/icons/',
  simpleicons: 'https://unpkg.com/simple-icons@latest/icons/'
};

export function iconUrl(icon: Pick<Icon, 'name' | 'source'>) {
  return `${ICON_BASE_URLS[icon.source]}${encodeURIComponent(icon.name)}.svg`;
}

async function fetchLucideIcon(iconName: string) {
  try {
    const response = await fetch(iconUrl({ name: iconName, source: 'lucide' }));
    if (response.ok) {
      return {
        svg: await response.text(),
        source: 'lucide',
        renderMode: 'stroke',
        needsViewportTransform: false
      };
    }

    return null;
  } catch (error) {
    console.error('Error fetching Lucide icon:', error);
    return null;
  }
}

async function fetchSimpleIcon(iconName: string) {
  try {
    const response = await fetch(iconUrl({ name: iconName, source: 'simpleicons' }));
    if (response.ok) {
      return {
        svg: await response.text(),
        source: 'simpleicons',
        renderMode: 'fill',
        needsViewportTransform: true
      };
    }

    return null;
  } catch (error) {
    console.error('Error fetching Simple Icon:', error);
    return null;
  }
}

// Fetch an icon from the given source, or try Lucide then Simple Icons when the source is unknown
export async function fetchIcon(iconName: string, source: IconSource | '' = '') {
  if (source === 'lucide') return fetchLucideIcon(iconName);
  if (source === 'simpleicons') return fetchSimpleIcon(iconName);

  let iconResult = null;
  iconResult = await fetchLucideIcon(iconName);
  if (iconResult) { return iconResult };

  iconResult = await fetchSimpleIcon(iconName);

  return iconResult;
}
