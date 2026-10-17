import { createContext, useCallback, useContext, useEffect, useRef, useState, type ReactNode } from "react";
import { api } from "./api";
import type { Frame } from "./types";

// World state is the latest server frame. Playing is view state: it only
// decides when the shell asks the server for another tick.

type Command = "state" | "tick" | "reset";

type Shell = {
  frame: Frame | null;
  error: string | null;
  playing: boolean;
  tick: () => void;
  reset: () => void;
  togglePlay: () => void;
};

const WorldContext = createContext<Shell | null>(null);

function wait(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

export function WorldProvider({ children }: { children: ReactNode }) {
  const [frame, setFrame] = useState<Frame | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [playing, setPlaying] = useState(false);
  const generation = useRef(0);
  const busy = useRef(false);

  const send = useCallback(async (command: Command) => {
    const gen = generation.current;
    while (busy.current) {
      await wait(40);
      if (command === "tick" && gen !== generation.current) return;
    }
    if (command === "tick" && gen !== generation.current) return;
    busy.current = true;
    try {
      const next = await api[command]();
      if (command === "tick" && gen !== generation.current) return;
      setFrame(next);
      setError(null);
    } catch (err) {
      if (command === "tick" && gen !== generation.current) return;
      setError(err instanceof Error ? err.message : "request failed");
      if (command === "tick") setPlaying(false);
    } finally {
      busy.current = false;
    }
  }, []);

  useEffect(() => {
    void send("state");
  }, [send]);

  useEffect(() => {
    if (!playing) return;
    const timer = setInterval(() => void send("tick"), 800);
    return () => clearInterval(timer);
  }, [playing, send]);

  const tick = useCallback(() => void send("tick"), [send]);

  const reset = useCallback(() => {
    generation.current += 1;
    setPlaying(false);
    void send("reset");
  }, [send]);

  const togglePlay = useCallback(() => {
    setPlaying((on) => !on);
  }, []);

  const shell: Shell = { frame, error, playing, tick, reset, togglePlay };
  return <WorldContext.Provider value={shell}>{children}</WorldContext.Provider>;
}

function useShell(): Shell {
  const shell = useContext(WorldContext);
  if (!shell) throw new Error("world hooks must be used inside WorldProvider");
  return shell;
}

export function useWorld() {
  const { frame, error, tick, reset } = useShell();
  return { frame, error, tick, reset };
}

export function useView() {
  const { playing, togglePlay } = useShell();
  return { playing, togglePlay };
}
