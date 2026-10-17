// Frame is the world snapshot served by cmd/web. The browser paints it
// and does not advance the simulation itself.

export type Parcel = {
  id: number;
  name: string;
  species: string;
  moles: number;
  x: number;
  y: number;
  temp: number;
};

export type Frame = {
  tick: number;
  mass: number;
  parcels: Parcel[];
};
