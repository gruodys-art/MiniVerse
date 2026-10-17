import { useView, useWorld } from "../world";

export function Header() {
  const { frame, error, tick, reset } = useWorld();
  const { playing, togglePlay } = useView();
  const ready = frame !== null;
  const stats = frame ? `tick ${frame.tick} · ${frame.mass.toFixed(3)} g` : "loading";

  return (
    <header>
      <h1>MiniVerse</h1>
      <p className="stats">{stats}</p>
      {error ? <p className="error">{error}</p> : null}
      <div className="actions">
        <button className="primary" type="button" onClick={togglePlay} disabled={!ready}>
          {playing ? "Pause" : "Play"}
        </button>
        <button type="button" onClick={tick} disabled={!ready}>
          Tick
        </button>
        <button type="button" onClick={reset} disabled={!ready}>
          Reset
        </button>
      </div>
    </header>
  );
}
