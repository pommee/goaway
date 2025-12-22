import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from "@/components/ui/dropdown-menu";
import { SunIcon } from "@phosphor-icons/react";
import { themes, themesConfig } from "./theme-context";
import { useTheme } from "./use-theme";

export function ModeToggle() {
  const { theme, setTheme } = useTheme();
  const CurrentIcon = themesConfig[theme]?.icon || SunIcon;

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" className="cursor-pointer">
          <CurrentIcon className="h-[1.2rem] w-[1.2rem] transition-transform hover:scale-110" />
          <span className="sr-only">Toggle theme</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-48">
        {themes.map((themeName) => {
          const { label, icon: Icon } = themesConfig[themeName];
          const isActive = theme === themeName;

          return (
            <DropdownMenuItem
              key={themeName}
              onClick={() => setTheme(themeName)}
              className="cursor-pointer flex items-center gap-2"
            >
              <Icon
                className={`h-4 w-4 transition-colors ${
                  isActive ? "text-primary" : "text-muted-foreground"
                }`}
              />
              <span className={isActive ? "font-medium text-primary" : ""}>
                {label}
              </span>
              {isActive && <span className="ml-auto text-primary">✓</span>}
            </DropdownMenuItem>
          );
        })}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
