import { useLayoutEffect, useRef } from "react";
import { draw } from "../draw";
import { useWorld } from "../world";

export function Stage() {
  const { frame } = useWorld();
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useLayoutEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas || !frame) return;
    const paint = () => draw(canvas, frame);
    paint();
    const parent = canvas.parentElement;
    if (!parent) return;
    const observer = new ResizeObserver(paint);
    observer.observe(parent);
    return () => observer.disconnect();
  }, [frame]);

  return (
    <section className="stage">
      <canvas ref={canvasRef} aria-hidden="true" />
      <p className="hint">Each circle is one pile. Size follows the amount. Color is the material. One tick is one second.</p>
    </section>
  );
}
