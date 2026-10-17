import { speciesColor } from "./palette";
import type { Frame } from "./types";

function reach(moles: number): number {
  return 0.5 + 0.5 * Math.sqrt(Math.max(moles, 0));
}

// draw paints one server frame onto the canvas. It does not fetch or store state.
export function draw(canvas: HTMLCanvasElement, frame: Frame): void {
  const parcels = frame.parcels;
  const dpr = window.devicePixelRatio || 1;
  const width = canvas.clientWidth;
  const height = canvas.clientHeight;
  canvas.width = Math.round(width * dpr);
  canvas.height = Math.round(height * dpr);
  const ctx = canvas.getContext("2d");
  if (!ctx) return;
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  ctx.clearRect(0, 0, width, height);

  let minX = Infinity;
  let maxX = -Infinity;
  for (const parcel of parcels) {
    minX = Math.min(minX, parcel.x - reach(parcel.moles));
    maxX = Math.max(maxX, parcel.x + reach(parcel.moles));
  }
  if (!Number.isFinite(minX)) {
    minX = 0;
    maxX = 10;
  }
  const spanX = Math.max(maxX - minX, 8);
  const spanY = 10;
  const pad = 28;
  const scale = Math.min((width - pad * 2) / spanX, (height - pad * 2) / spanY);
  const originX = pad + (width - pad * 2 - spanX * scale) / 2;
  const toX = (x: number) => originX + (x - minX) * scale;
  const toY = (y: number) => height / 2 - y * scale;

  const ordered = parcels.slice().sort((a, b) => b.moles - a.moles);
  for (const parcel of ordered) {
    const radius = Math.max(reach(parcel.moles) * scale, 8);
    ctx.beginPath();
    ctx.fillStyle = speciesColor(parcel.species);
    ctx.arc(toX(parcel.x), toY(parcel.y), radius, 0, Math.PI * 2);
    ctx.fill();
    ctx.lineWidth = 1;
    ctx.strokeStyle = "rgba(28, 36, 48, 0.35)";
    ctx.stroke();
  }
}
