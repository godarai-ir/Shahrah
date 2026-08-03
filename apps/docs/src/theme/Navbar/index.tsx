import {
  applyThemePreference,
  getBrowserThemePreference,
  migrateBrowserThemePreference,
  setBrowserThemePreference,
  subscribeToThemePreference,
} from '@shahrah/theme-sync';
import type {ThemePreference} from '@shahrah/theme-sync';
import {useEffect, useRef, type ReactNode} from 'react';
import OriginalNavbar from '@theme-original/Navbar';
import Link from '@docusaurus/Link';
import {useColorMode} from '@docusaurus/theme-common';

const themeReadOptions = {
  defaultTheme: 'system' as const,
  legacyStorageKeys: ['theme', 'shahrah-theme', 'vite-ui-theme'],
};

function SharedThemeSync(): null {
  const {colorModeChoice, setColorMode} = useColorMode();
  const didMount = useRef(false);

  useEffect(() => {
    const syncFromSharedPreference = () => {
      const preference = getBrowserThemePreference(themeReadOptions);
      setColorMode(preference === 'system' ? null : preference);
      applyThemePreference(preference);
    };

    migrateBrowserThemePreference(themeReadOptions);
    syncFromSharedPreference();
    return subscribeToThemePreference(syncFromSharedPreference);
  }, [setColorMode]);

  useEffect(() => {
    if (!didMount.current) {
      didMount.current = true;
      return;
    }

    const preference: ThemePreference = colorModeChoice ?? 'system';
    setBrowserThemePreference(preference);
    applyThemePreference(preference);
  }, [colorModeChoice]);

  return null;
}

export default function NavbarWrapper(props: Record<string, unknown>): ReactNode {
  return (
    <div className="navbar-with-inline-search">
      <SharedThemeSync />
      <OriginalNavbar {...props} />
      <Link className="navbar-inline-search" to="/search">
        جستجو
      </Link>
    </div>
  );
}
