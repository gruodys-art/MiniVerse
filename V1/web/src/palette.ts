// Species color is presentation. It is not part of the world snapshot.

const colors: Record<string, string> = {
  water: "#3d8bfd",
  ice: "#d7eeff",
  steam: "#c5d0dc",
  salt: "#f7f4ee",
  dissolved_salt: "#7ec8e3",
  acid: "#e85d4c",
  base: "#3cba8a",
  iron: "#8d8f96",
  oxygen: "#9ad7c8",
  iron_oxide: "#c4652d",
};

export function speciesColor(species: string): string {
  return colors[species] ?? "#999";
}
