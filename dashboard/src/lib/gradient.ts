function hexToRgb(hex: string) {
    hex = hex.replace(/^#/, '');
    if (hex.length === 3) {
        hex = hex.split('').map(x => x + x).join('');
    }

    const bigint = parseInt(hex, 16);
    return [(bigint >> 16) & 255, (bigint >> 8) & 255, bigint & 255];
}

function rgbToHex(r: number, g: number, b: number) {
    return `#${((1 << 24) + (r << 16) + (g << 8) + b).toString(16).slice(1).toUpperCase()}`;
}

export function blendColors(color1: string, color2: string, percentage: number) {
    const rgb1 = hexToRgb(color1);
    const rgb2 = hexToRgb(color2);

    // Interpolate each RGB component
    const r = Math.round(rgb1[0] + (rgb2[0] - rgb1[0]) * percentage);
    const g = Math.round(rgb1[1] + (rgb2[1] - rgb1[1]) * percentage);
    const b = Math.round(rgb1[2] + (rgb2[2] - rgb1[2]) * percentage);

    return rgbToHex(r, g, b);
}