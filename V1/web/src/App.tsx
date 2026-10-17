import { Header } from "./components/Header";
import { ParcelTable } from "./components/ParcelTable";
import { Stage } from "./components/Stage";
import { WorldProvider } from "./world";

export function App() {
  return (
    <WorldProvider>
      <Header />
      <main>
        <Stage />
        <ParcelTable />
      </main>
    </WorldProvider>
  );
}
