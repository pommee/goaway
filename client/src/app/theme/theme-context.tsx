import { createContext } from "react";
import {
  MoonIcon,
  SunIcon,
  DropIcon,
  TreeIcon,
  SparkleIcon,
  FlowerIcon,
  MonitorIcon,
  BookOpenIcon,
  CloudIcon,
  FireIcon,
  GameControllerIcon,
  LeafIcon,
  LightningIcon,
  MoonStarsIcon,
  PaintBrushIcon,
  PaletteIcon,
  SnowflakeIcon,
  WavesIcon,
  MountainsIcon
} from "@phosphor-icons/react";

export const themes = [
  "system",
  "light",
  "dark",
  "blue",
  "sunset",
  "forest",
  "purple",
  "rose",
  "cyberpunk",
  "cotton-candy",
  "midnight",
  "amber",
  "emerald",
  "slate",
  "ocean",
  "crimson",
  "nord",
  "vintage",
  "neon",
  "dusk",
  "pastel"
] as const;

export const themesConfig: Record<
  Theme,
  { label: string; icon: React.ComponentType<{ className?: string }> }
> = {
  light: { label: "Light", icon: SunIcon },
  dark: { label: "Dark", icon: MoonIcon },
  blue: { label: "Blue Ocean", icon: DropIcon },
  sunset: { label: "Sunset", icon: SunIcon },
  forest: { label: "Forest", icon: TreeIcon },
  purple: { label: "Purple Dream", icon: SparkleIcon },
  rose: { label: "Rose", icon: FlowerIcon },
  system: { label: "System", icon: MonitorIcon },
  cyberpunk: { label: "Cyberpunk", icon: GameControllerIcon },
  "cotton-candy": { label: "Cotton Candy", icon: CloudIcon },
  midnight: { label: "Midnight", icon: MoonStarsIcon },
  amber: { label: "Amber", icon: SunIcon },
  emerald: { label: "Emerald", icon: LeafIcon },
  slate: { label: "Slate", icon: PaletteIcon },
  ocean: { label: "Ocean", icon: WavesIcon },
  crimson: { label: "Crimson", icon: FireIcon },
  nord: { label: "Nord", icon: SnowflakeIcon },
  vintage: { label: "Vintage", icon: BookOpenIcon },
  neon: { label: "Neon", icon: LightningIcon },
  dusk: { label: "Dusk", icon: MountainsIcon },
  pastel: { label: "Pastel", icon: PaintBrushIcon }
};

export type Theme = (typeof themes)[number];

export type ThemeProviderState = {
  theme: Theme;
  setTheme: (theme: Theme) => void;
};

export const initialState: ThemeProviderState = {
  theme: "system",
  setTheme: () => null
};

export const ThemeProviderContext =
  createContext<ThemeProviderState>(initialState);
