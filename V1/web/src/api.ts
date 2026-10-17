import type { Frame } from "./types";

async function request(path: string, method: "GET" | "POST"): Promise<Frame> {
  const response = await fetch(path, { method });
  if (!response.ok) {
    throw new Error(response.statusText);
  }
  return response.json() as Promise<Frame>;
}

export const api = {
  state: () => request("/api/state", "GET"),
  tick: () => request("/api/tick", "POST"),
  reset: () => request("/api/reset", "POST"),
};
