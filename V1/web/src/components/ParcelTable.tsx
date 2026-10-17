import { speciesColor } from "../palette";
import { useWorld } from "../world";

function celsius(kelvin: number) {
  return kelvin - 273.15;
}

export function ParcelTable() {
  const { frame } = useWorld();

  return (
    <section className="list">
      <table>
        <thead>
          <tr>
            <th>Material</th>
            <th>Moles</th>
            <th>Temperature</th>
            <th>Position</th>
          </tr>
        </thead>
        <tbody>
          {(frame?.parcels ?? []).map((parcel) => (
            <tr key={parcel.id}>
              <td>
                <span className="swatch" style={{ background: speciesColor(parcel.species) }} />
                {parcel.name}
              </td>
              <td>{parcel.moles.toFixed(3)}</td>
              <td>
                {parcel.temp.toFixed(1)} K ({celsius(parcel.temp).toFixed(1)} °C)
              </td>
              <td>
                {parcel.x.toFixed(1)}, {parcel.y.toFixed(1)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}
