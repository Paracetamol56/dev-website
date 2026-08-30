async function fetchLucideIcon(iconName: string) {
  try {
    const response = await fetch(`https://unpkg.com/lucide-static@latest/icons/${iconName}.svg`);
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
    const response = await fetch(`https://unpkg.com/simple-icons@latest/icons/${iconName}.svg`);
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

// Function to fetch icon based on source
export async function fetchIcon(iconName: string) {
  let iconResult = null;
  iconResult = await fetchLucideIcon(iconName);
  if (iconResult) { return iconResult };

  iconResult = await fetchSimpleIcon(iconName);

  return iconResult;
}
