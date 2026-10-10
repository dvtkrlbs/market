import { Theme } from "@astryxdesign/core/theme";
import { neutralTheme } from "@astryxdesign/theme-neutral/built";
import type { ReactNode } from "react";

export const theme = neutralTheme;
export const themeMode = "system";

export function AppTheme({ children }: { children: ReactNode }) {
  return (
    <Theme theme={theme} mode={themeMode}>
      {children}
    </Theme>
  );
}
